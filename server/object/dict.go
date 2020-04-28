package object

type Dict struct {
	data map[interface{}]*Object
}

// 然后这里有一些dict的方法，来控制这些情况
// 比如传入object，我就知道底层数据
// 不需要做那么多事情
// 然后这里也封装一些跟redis类似的逻辑
// ok我准备吧所有的逻辑在回顾一遍
// 然后吧hash和ziplist重写调
// 并且测试下
// 如果这个写好后就可以开始写时间循环
// 不对我应该先时间循环
// 在AOF和ROF
// 然后在回顾再从新梳理整个事情
// 应该是这个顺序
// ok时间循环搞起