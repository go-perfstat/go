# go-perfstat

Record/calculate performance statistic for aggregation period and grand total.  
Once per aggregation time period flush samples somethere for Graphana to pick-up

### Create an instance per component

    perf := perfstat.ForName("domain")

### Use in a method

    t := perf.Start()
    ...
    perf.Stop(t)

### Print all stats before exit

    fmt.Printf("%-15s %-20s %-8s %-8s %-8s\n", "Type", "Name", "Min(ms)", "Avg(ms)", "Max(ms)")
    fmt.Println("-------------------------------------------------------------")
    ForEachOrdered(perfstat.GetAll(), func(typ string, innerMap map[string]*perfstat.Stat) {
        ForEachOrdered(innerMap, func(name string, st *perfstat.Stat) {
            fmt.Printf("%-15s %-20s %-8.3f %-8.3f %-8.3f\n",
                typ, name, st.GetMinTimeMs(), st.GetAvgTimeMs(), st.GetMaxTimeMs())
        })
    })

    func ForEachOrdered[V any](m map[string]V, fn func(key string, value V)) {
        keys := make([]string, 0, len(m))
        for k := range m {
            keys = append(keys, k)
        }
        sort.Strings(keys)
        for _, k := range keys {
            fn(k, m[k])
        }
    }