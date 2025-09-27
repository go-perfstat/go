# go-perfstat

Record/calculate performance statistic for aggregation period and grand total.  
Expose metrics using prometheus.

### Create an instance per component

    perf := perfstat.ForTypeName("package.Type", "Method")

### Use in a method

    t := perf.Start()
    ...
    perf.Stop(t)


### Register in prometheus

	prometheus.MustRegister(NewPerfStatMetricsCollector())

> Default aggregation period is 5s

### Print all stats before exit

	func printPerfStat() {
		fmt.Printf("%-70s %10s %10s %10s %10s %10s %5s\n", "Type/Name", "Min(ms)", "Avg(ms)", "Max(ms)", "Total(s)", "Leaps", "Peers")
		fmt.Println(strings.Repeat("-", 130))
		ForEachOrdered(perfstat.GetAll(), func(typ string, innerMap map[string]*perfstat.Stat) {
			ForEachOrdered(innerMap, func(name string, st *perfstat.Stat) {
				fmt.Printf("%-70s %10.3f %10.3f %10.3f %10.3f %10d %5d\n",
					strings.Join([]string{typ, name}, "."), st.GetMinTimeMs(), st.GetAvgTimeMs(), st.GetMaxTimeMs(), 
					math.Round(st.GetTotalTimeMs()/1000), st.GetLeapsCount(), st.GetPeersCount())
			})
		})
	}

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