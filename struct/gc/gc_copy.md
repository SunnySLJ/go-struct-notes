## gc 的调用时机

整个 gc 流程的入口是 gcStart，gcStart 的调用方为:

```mermaid
graph LR
mallocgc --> shouldhelpgc
shouldhelpgc --> |no|return
shouldhelpgc --> |yes|gc_condition_satisfied
gc_condition_satisfied --> |no|return
gc_condition_satisfied --> |yes|gcStart
runtime.GC --> gcStart
init --> forcegchelper
forcegchelper --> gcStart
```

其实就是 mallocgc，forcegchelper，runtime.GC 这三个入口。

* mallocgc，分配堆内存时触发，会检查当前是否满足触发 gc 的条件，如果触发，那么进入 gcStart。
* forcegchelper，在 forcegchelper 中会把 forcegc.g 这个全局对象的运行 g 挂起。sysmon 会调用 test 检查上次触发 gc 的时间到当前时间是否已经经过了
* forcegcperiod 长的时间，如果已经超过，那么就会将 forcegc.g 注入到 globrunq。这样会在该 g 被调度到的时候触发 gc。
* runtime.GC 是由用户主动触发的，相当于强制触发 GC。

## gcTrigger 和 gc 条件检查

```go
// A gcTrigger is a predicate for starting a GC cycle. Specifically,
// it is an exit condition for the _GCoff phase.
type gcTrigger struct {
    kind gcTriggerKind
    now  int64  // gcTriggerTime: current time
    n    uint32 // gcTriggerCycle: cycle number to start
}

type gcTriggerKind int

const (
    // 表示应该无条件地开始 GC，不管外部任何参数
    // 即使 GOGC 设置为 off，或者当前已经在进行 GC 进行中
    gcTriggerAlways gcTriggerKind = iota

    // 该枚举值表示 GC 会在 heap 大小达到 controller 计算出的阈值之后开始
    gcTriggerHeap

    // 表示从上一次 GC 之后，经过了 forcegcperiod 纳秒
    // 基于时间触发本次 GC
    gcTriggerTime

    // gcTriggerCycle indicates that a cycle should be started if
    // we have not yet started cycle number gcTrigger.n (relative
    // to work.cycles).
    gcTriggerCycle
)
```

### mallocgc 中的 trigger 类型是 gcTriggerHeap;

```go
    if shouldhelpgc {
        if t := (gcTrigger{kind: gcTriggerHeap}); t.test() {
            gcStart(gcBackgroundMode, t)
        }
    }
```

在 mallocgc 中进行 gc 可以防止内存分配过快，导致 GC 回收不过来。

### runtime.GC 中使用的是 gcTriggerCycle;

```go
    // We're now in sweep N or later. Trigger GC cycle N+1, which
    // will first finish sweep N if necessary and then enter sweep
    // termination N+1.
    gcStart(gcBackgroundMode, gcTrigger{kind: gcTriggerCycle, n: n + 1})
```

### forcegchelper 中使用的是 gcTriggerTime;

```go
    // Time-triggered, fully concurrent.
    gcStart(gcBackgroundMode, gcTrigger{kind: gcTriggerTime, now: nanotime()})
```

### sysmon 中检查时使用的也是 gcTriggerTime

这里和前面三条不一样的是，这种情况下如果 trigger.test 返回 true，会使用 forcegchelper 所在的 g 来执行 gcStart，具体做法就是上面提到的把 forcegc.g 注入到全局 runq;

```go
    // check if we need to force a GC
    if t := (gcTrigger{kind: gcTriggerTime, now: now}); t.test() && atomic.Load(&forcegc.idle) != 0 {
        lock(&forcegc.lock)
        forcegc.idle = 0
        forcegc.g.schedlink = 0
        injectglist(forcegc.g)
        unlock(&forcegc.lock)
    }
```

### trigger.test, gc 条件检查

```go
// test returns true if the trigger condition is satisfied, meaning
// that the exit condition for the _GCoff phase has been met. The exit
// condition should be tested when allocating.
func (t gcTrigger) test() bool {
    if !memstats.enablegc || panicking != 0 {
        return false
    }
    if t.kind == gcTriggerAlways {
        return true
    }
    if gcphase != _GCoff {
        return false
    }
    switch t.kind {
    case gcTriggerHeap:
        // Non-atomic access to heap_live for performance. If
        // we are going to trigger on this, this thread just
        // atomically wrote heap_live anyway and we'll see our
        // own write.
        return memstats.heap_live >= memstats.gc_trigger
    case gcTriggerTime:
        if gcpercent < 0 {
            return false
        }
        lastgc := int64(atomic.Load64(&memstats.last_gc_nanotime))
        return lastgc != 0 && t.now-lastgc > forcegcperiod
    case gcTriggerCycle:
        // t.n > work.cycles, but accounting for wraparound.
        return int32(t.n-work.cycles) > 0
    }
    return true
}
```

## gcStart

```go
// 该函数会将 GC 状态从 _GCoff 切换到 _GCmark(if !mode.stwMark)
// 或者 _GCmarktermination(if mode.stwMark)
// 开始执行 sweep termination 或者 GC 初始化
//
// 函数可能会在一些情况下未进行状态变更就返回
// 比如在 system stack 中被调用，或者 locks 被别人持有
func gcStart(mode gcMode, trigger gcTrigger) {
    // Since this is called from malloc and malloc is called in
    // the guts of a number of libraries that might be holding
    // locks, don't attempt to start GC in non-preemptible or
    // potentially unstable situations.
    mp := acquirem()
    if gp := getg(); gp == mp.g0 || mp.locks > 1 || mp.preemptoff != "" {
        releasem(mp)
        return
    }
    releasem(mp)
    mp = nil

    // Pick up the remaining unswept/not being swept spans concurrently
    //
    // This shouldn't happen if we're being invoked in background
    // mode since proportional sweep should have just finished
    // sweeping everything, but rounding errors, etc, may leave a
    // few spans unswept. In forced mode, this is necessary since
    // GC can be forced at any point in the sweeping cycle.
    //
    // We check the transition condition continuously here in case
    // this G gets delayed in to the next GC cycle.
    for trigger.test() && gosweepone() != ^uintptr(0) {
        sweep.nbgsweep++
    }

    // 进行 GC 初始化和 sweep termination
    semacquire(&work.startSema)
    // 在 transition lock 下再检查一次 transition 条件
    if !trigger.test() {
        semrelease(&work.startSema)
        return
    }

    // For stats, check if this GC was forced by the user.
    work.userForced = trigger.kind == gcTriggerAlways || trigger.kind == gcTriggerCycle

    // In gcstoptheworld debug mode, upgrade the mode accordingly.
    // We do this after re-checking the transition condition so
    // that multiple goroutines that detect the heap trigger don't
    // start multiple STW GCs.
    if mode == gcBackgroundMode {
        if debug.gcstoptheworld == 1 {
            mode = gcForceMode
        } else if debug.gcstoptheworld == 2 {
            mode = gcForceBlockMode
        }
    }

    // 获取全局锁，让别人都停下 stw
    semacquire(&worldsema)

    // 在 background 模式下
    // 启动所有标记 worker
    if mode == gcBackgroundMode {
        gcBgMarkStartWorkers()
    }

    gcResetMarkState()

    work.stwprocs, work.maxprocs = gomaxprocs, gomaxprocs
    if work.stwprocs > ncpu {
        // This is used to compute CPU time of the STW phases,
        // so it can't be more than ncpu, even if GOMAXPROCS is.
        work.stwprocs = ncpu
    }
    work.heap0 = atomic.Load64(&memstats.heap_live)
    work.pauseNS = 0
    work.mode = mode

    now := nanotime()
    work.tSweepTerm = now
    work.pauseStart = now

    systemstack(stopTheWorldWithSema)
    // Finish sweep before we start concurrent scan.
    systemstack(func() {
        finishsweep_m()
    })

    // 清除全局的 :
    // 1. sudogcache(sudog 数据结构的链表)
    // 2. deferpool(defer struct 的链表的数组)
    // 3. sync.Pool
    clearpools()

    work.cycles++
    if mode == gcBackgroundMode { // 尽量多地提高并发度
        gcController.startCycle()
        work.heapGoal = memstats.next_gc

        // 进入并发标记阶段，并让 write barriers 开始生效
        //
        // Because the world is stopped, all Ps will
        // observe that write barriers are enabled by
        // the time we start the world and begin
        // scanning.
        //
        // Write barriers must be enabled before assists are
        // enabled because they must be enabled before
        // any non-leaf heap objects are marked. Since
        // allocations are blocked until assists can
        // happen, we want enable assists as early as
        // possible.
        setGCPhase(_GCmark)

        gcBgMarkPrepare() // Must happen before assist enable.
        gcMarkRootPrepare()

        // Mark all active tinyalloc blocks. Since we're
        // allocating from these, they need to be black like
        // other allocations. The alternative is to blacken
        // the tiny block on every allocation from it, which
        // would slow down the tiny allocator.
        gcMarkTinyAllocs()

        // At this point all Ps have enabled the write
        // barrier, thus maintaining the no white to
        // black invariant. Enable mutator assists to
        // put back-pressure on fast allocating
        // mutators.
        atomic.Store(&gcBlackenEnabled, 1)

        // 协助标记的 g 和 worker 在 start the world 之后就可以开始工作了
        gcController.markStartTime = now

        // 并发标记
        systemstack(func() {
            now = startTheWorldWithSema(trace.enabled)
        })
        work.pauseNS += now - work.pauseStart
        work.tMark = now
    } else {
        t := nanotime()
        work.tMark, work.tMarkTerm = t, t
        work.heapGoal = work.heap0

        // 进行 mark termination 阶段的工作，会 restart the world
        gcMarkTermination(memstats.triggerRatio)
    }

    semrelease(&work.startSema)
}
```

## runtime.GC

```go
// GC runs a garbage collection and blocks the caller until the
// garbage collection is complete. It may also block the entire
// program.
func GC() {
    // We consider a cycle to be: sweep termination, mark, mark
    // termination, and sweep. This function shouldn't return
    // until a full cycle has been completed, from beginning to
    // end. Hence, we always want to finish up the current cycle
    // and start a new one. That means:
    //
    // 1. In sweep termination, mark, or mark termination of cycle
    // N, wait until mark termination N completes and transitions
    // to sweep N.
    //
    // 2. In sweep N, help with sweep N.
    //
    // At this point we can begin a full cycle N+1.
    //
    // 3. Trigger cycle N+1 by starting sweep termination N+1.
    //
    // 4. Wait for mark termination N+1 to complete.
    //
    // 5. Help with sweep N+1 until it's done.
    //
    // This all has to be written to deal with the fact that the
    // GC may move ahead on its own. For example, when we block
    // until mark termination N, we may wake up in cycle N+2.

    gp := getg()

    // Prevent the GC phase or cycle count from changing.
    lock(&work.sweepWaiters.lock)
    n := atomic.Load(&work.cycles)
    if gcphase == _GCmark {
        // Wait until sweep termination, mark, and mark
        // termination of cycle N complete.
        gp.schedlink = work.sweepWaiters.head
        work.sweepWaiters.head.set(gp)
        goparkunlock(&work.sweepWaiters.lock, "wait for GC cycle", traceEvGoBlock, 1)
    } else {
        // We're in sweep N already.
        unlock(&work.sweepWaiters.lock)
    }

    // We're now in sweep N or later. Trigger GC cycle N+1, which
    // will first finish sweep N if necessary and then enter sweep
    // termination N+1.
    gcStart(gcBackgroundMode, gcTrigger{kind: gcTriggerCycle, n: n + 1})

    // Wait for mark termination N+1 to complete.
    lock(&work.sweepWaiters.lock)
    if gcphase == _GCmark && atomic.Load(&work.cycles) == n+1 {
        gp.schedlink = work.sweepWaiters.head
        work.sweepWaiters.head.set(gp)
        goparkunlock(&work.sweepWaiters.lock, "wait for GC cycle", traceEvGoBlock, 1)
    } else {
        unlock(&work.sweepWaiters.lock)
    }

    // Finish sweep N+1 before returning. We do this both to
    // complete the cycle and because runtime.GC() is often used
    // as part of tests and benchmarks to get the system into a
    // relatively stable and isolated state.
    for atomic.Load(&work.cycles) == n+1 && gosweepone() != ^uintptr(0) {
        sweep.nbgsweep++
        Gosched()
    }

    // Callers may assume that the heap profile reflects the
    // just-completed cycle when this returns (historically this
    // happened because this was a STW GC), but right now the
    // profile still reflects mark termination N, not N+1.
    //
    // As soon as all of the sweep frees from cycle N+1 are done,
    // we can go ahead and publish the heap profile.
    //
    // First, wait for sweeping to finish. (We know there are no
    // more spans on the sweep queue, but we may be concurrently
    // sweeping spans, so we have to wait.)
    for atomic.Load(&work.cycles) == n+1 && atomic.Load(&mheap_.sweepers) != 0 {
        Gosched()
    }

    // Now we're really done with sweeping, so we can publish the
    // stable heap profile. Only do this if we haven't already hit
    // another mark termination.
    mp := acquirem()
    cycle := atomic.Load(&work.cycles)
    if cycle == n+1 || (gcphase == _GCmark && cycle == n+2) {
        mProf_PostSweep()
    }
    releasem(mp)
}
```

## FAQ

为什么需要 debug.FreeOsMemory 才能释放堆空间。

concurrent sweep 和普通后台sweep有什么区别，分别在什么时间触发

stw的时候在干什么

优化gc主要是在优化什么

gc时间，stw时间和响应延迟之间是什么关系

宏观来看gc划分为多少个阶段


<img width="330px"  src="https://xargin.com/content/images/2021/05/wechat.png">
