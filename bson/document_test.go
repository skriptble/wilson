package bson

func ExampleDocumentWalk() {
	d := new(Document)
	list := make([][]string, 0)
	d.Walk(true, func(prefix []string, e *Element) {
		if e.Key() == "ui" {
			l := make([]string, len(prefix), len(prefix)+1)
			copy(l, prefix)
			l = append(l, "ui")
			list = append(list, l)
		}
	})
	for _, l := range list {
		err := d.Delete(l...).Err()
		if err != nil {
			panic(err)
		}
	}
}
