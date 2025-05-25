package leetcode

//       int                                        12345(6)     index集合
//      /   \
//     e     r                                 1234         5(6)
//    / \      \
//    r  n      u                        123(6)         4           5
//  /  \  \   /
// v   n  sion                        1     23(6)             45
//  \ / \
//   al  et                              12(6)     3
//
// 第一次遍历 生成图
// 第二次遍历 生成所有环 排序
// 例:
// 123456
// 1234 [-5, -6]
// 1236 [-5, -6][-4, 6]
// 1 [-5, -6][-4, 6][-2, -3, -6]
// 126 [-5, -6][-4, 6][-2, -3, -6][2, 6]
// []{[[-5, -6], [-4, 6], [-2, -3, -6], [2, 6]], }
// 路径切片中唯一不存在的下标 1 因此该路径是1节点的路径
// 若该节点存在子节点不包含1 说明1在该节点终结
// 生成所有终结节点的所有环 struct{height:2, node:1, other:2, sx:3, sy:1, ex:4, ey:1} 根据所有height排序
// 从最长的height生成缩写 放入 map[node(int)]struct
// node == map node || node == map other 则跳过
// 第三次遍历 生成字符串 sx到ex 替换为数字
//
// 开什么玩笑 要吐了

func wordsAbbreviation(words []string) []string {
	nodeFigure := make([][]*node, 0, 8)
	roots := initFigureStructure(words, &nodeFigure)
	heights := make([]*height, 0)
	for _, root := range roots {
		heights = append(heights, getHeights(nil, root, make(map[int]struct{}), make([][]int, 0), make(map[int]int), root.index)...)
	}
	m := make(map[int]struct{})
	for _, h := range quickSort(heights) {
		if _, ok := m[**h.curr]; ok {
			continue
		}
		if _, ok := m[h.other]; ok {
			continue
		}
		m[**h.curr] = struct{}{}
		if h.ex-h.sx > 1 {
			index := **h.curr - 1
			runes := []rune(words[index])
			words[index] = string(append(append(runes[:h.sx+1], rune(h.ex-h.sx+47)), runes[h.ex:]...))
		}
	}
	return words
}

// initFigureStructure 生成图
func initFigureStructure(words []string, nodeFigure *[][]*node) []*node {
	for i, word := range words {
		var current *node
		for j, rnu := range word {
			if j > len(*nodeFigure)-1 {
				*nodeFigure = append(*nodeFigure, make([]*node, 0, 4))
			}
			// TODO rune
			rn := string(rnu)

			// head
			if current == nil {
				find := false
				for _, root := range (*nodeFigure)[0] {
					if root != nil && root.rn == rn {
						root.index[i+1] = struct{}{}
						current = root
						find = true
						break
					}
				}
				if !find {
					index := make(map[int]struct{})
					index[i+1] = struct{}{}
					current = &node{
						height: j,
						rn:     rn,
						index:  index,
					}
					(*nodeFigure)[0] = append((*nodeFigure)[0], current)
				}
				continue
			}

			if current.childMap == nil {
				current.childMap = make(map[string]*node)
			}

			// self child
			if child, ok := current.childMap[rn]; ok {
				current = child
				child.index[i+1] = struct{}{}
				continue
			}

			// brothers child
			find := false
			for _, brother := range (*nodeFigure)[j] {
				if brother.childMap == nil {
					continue
				}
				if child := brother.childMap[rn]; child != nil {
					child.index[i+1] = struct{}{}
					current.childMap[rn] = child
					current = child
					find = true
					break
				}
			}
			if find {
				continue
			}

			// child not found
			index := make(map[int]struct{})
			index[i+1] = struct{}{}
			child := &node{
				height: j,
				rn:     rn,
				index:  index,
			}
			(*nodeFigure)[j] = append((*nodeFigure)[j], child)
			current.childMap[rn] = child
			current = child
		}
	}
	return (*nodeFigure)[0]
}

// getHeights 生成所有环
func getHeights(curr **int, root *node, parentIndex map[int]struct{}, slices [][]int, idxMap map[int]int, allIndex map[int]struct{}) (heights []*height) {
	x := root.height
	slices = append(slices, make([]int, 0))
	for idx := range parentIndex {
		if _, ok := root.index[idx]; ok {
			continue
		}
		slices[x] = append(slices[x], -idx)
	}
	for idx := range root.index {
		if _, ok := parentIndex[idx]; ok {
			continue
		}
		slices[x] = append(slices[x], idx)
	}

	// all contains flag
	contains := true
	for idx := range allIndex {
		_, contains = idxMap[idx]
		if !contains {
			tp := &idx
			curr = &tp
			break
		}
	}
	// end
	if contains {
		return
	}
	// build heights
	res := make([]*height, 0)
	if x > 0 {
		for y, idx := range slices[x] {
			if idx < 0 {
				idxMap[-idx] = (x << 16) | y
			} else {
				xy, ok := idxMap[idx]
				//TODO
				if !ok {
					panic("out without in")
				}
				sx := 0b111111111111111111111111000000000000000000000000 & xy
				sy := xy >> 16
				res = append(res, &height{
					height: x - sx,
					curr:   curr,
					other:  idx,
					sx:     sx,
					sy:     sy,
					ex:     x,
					ey:     y,
				})
			}
		}
	}

	// build height
	for _, child := range root.childMap {
		newSlices := make([][]int, len(slices))
		copy(newSlices, slices)
		heights = append(res, getHeights(curr, child, root.index, newSlices, idxMap, allIndex)...)
	}
	return heights

}

// quickSort 快排
func quickSort(arr []*height) []*height {
	if len(arr) < 2 {
		return arr
	}

	pivot := arr[0]
	left, right := []*height{}, []*height{}

	for _, v := range arr[1:] {
		if v.height >= pivot.height {
			left = append(left, v)
		} else {
			right = append(right, v)
		}
	}

	return append(append(quickSort(left), pivot), quickSort(right)...)
}

type node struct {
	rn       string
	height   int
	index    map[int]struct{}
	childMap map[string]*node
}

type height struct {
	height int
	curr   **int
	other  int
	sx     int
	sy     int
	ex     int
	ey     int
}
