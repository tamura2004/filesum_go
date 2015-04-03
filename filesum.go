package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	//"strings"
)

// 情報格納用の木構造
type SubDir []*Dir

type Dir struct {
	path     string
	base     string
	size     int64
	depth    int
	parent   *Dir
	children SubDir
}

func (p SubDir) Len() int {
	return len(p)
}

func (p SubDir) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p SubDir) Less(i, j int) bool {
	return p[i].size > p[j].size
}

var limit int64

func main() {
	var path string
	flag.StringVar(&path, "d", ".", "root directory")
	flag.Int64Var(&limit, "l", 10000000, "file size filter")
	flag.Parse()

	path, err := filepath.Abs(path)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("invalid pathname: %s\n")
	}

	root := new(Dir)
	root.path = path
	root.base = filepath.Base(path)
	root.depth = 0
	root.Collect()
	root.Display("", true)
}

// ファイルサイズを適切な単位に変換する
func formatByUnit(size int64) string {
	unit := "byte"
	units := []string{"KB", "MB", "GB"}

	for _, u := range units {
		if size >= 1024 {
			size /= 1024
			unit = u
		}
	}
	return fmt.Sprintf("%d %s", size, unit)
}

// Collectしたツリー情報を表示
func (d *Dir) Display(indent string, isLast bool) {
	if isLast {
		fmt.Printf("%s%s%s .. %s\n", indent, "└", d.base, formatByUnit(d.size))
		indent += "  "
	} else {
		fmt.Printf("%s%s%s .. %s\n", indent, "├", d.base, formatByUnit(d.size))
		indent += "│"
	}

	sort.Sort(d.children)
	for i, c := range d.children {
		if i == len(d.children)-1 {
			c.Display(indent, true)
			if c.children == nil {
				fmt.Println(indent)
			}
		} else {
			c.Display(indent, false)
		}
	}
}

// ファイルサイズを取得して木構造に格納
func (d *Dir) Collect() int64 {
	f, err := os.Open(d.path)
	if err != nil {
		//fmt.Println(err)
		//fmt.Printf("cannot open %s\n", path)
		return 0
	}

	fis, err := f.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("cannot get infomation of %s\n", d.path)
		return 0
	}

	for _, fi := range fis {
		if fi.Mode()&os.ModeSymlink != 0 {
			// do nothing if symbolic link
		} else if fi.IsDir() {
			c := new(Dir)
			c.base = fi.Name()
			c.path = filepath.Join(d.path, c.base)
			c.depth = d.depth + 1
			c.parent = d
			size := c.Collect()
			d.size += size
			if size > limit {
				d.children = append(d.children, c)
			}
		} else {
			d.size += fi.Size()
		}
	}

	return d.size
}
