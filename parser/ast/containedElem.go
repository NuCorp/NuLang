package ast

type ContainedElem interface {
	GetExpr() Expr
	IsNamed() bool
	IsDeep() bool
	IsDestructured() bool
}

type OrderedContainedElem struct {
	Expr
}

func (o OrderedContainedElem) GetExpr() Expr {
	return o.Expr
}

func (o OrderedContainedElem) IsNamed() bool {
	return false
}

func (o OrderedContainedElem) IsDeep() bool { return false }

func (o OrderedContainedElem) IsDestructured() bool {
	return false
}

type NamedContainedElem struct {
	Name DotIdent
	Expr Expr // may be nil
	Deep bool
}

func (n NamedContainedElem) GetExpr() Expr {
	if n.Expr == nil {
		return n.Name
	}

	return n.Expr
}

func (NamedContainedElem) IsNamed() bool {
	return true
}

func (NamedContainedElem) IsDestructured() bool {
	return false
}

func (n NamedContainedElem) IsDeep() bool { return n.Deep }

type DestructuredContainedElem[T ContainedElem] struct {
	ContainedElem T
}

func (d DestructuredContainedElem[T]) GetExpr() Expr {
	return d.ContainedElem.GetExpr()
}

func (d DestructuredContainedElem[T]) IsNamed() bool {
	return d.ContainedElem.IsNamed()
}

func (d DestructuredContainedElem[T]) IsDeep() bool { return d.ContainedElem.IsDeep() }

func (d DestructuredContainedElem[T]) IsDestructured() bool {
	return true
}

type containedElem[Ordered, Named, Destructured ContainedElem] struct {
	Ordered      *Ordered
	Named        *Named
	Destructured *Destructured
}

func (c *containedElem[Ordered, Named, Destructured]) Set(elem ContainedElem) {
	var set func()

	defer func() {
		if set == nil {
			return
		}

		*c = containedElem[Ordered, Named, Destructured]{}
		set()
	}()

	switch elem := elem.(type) {
	case Ordered:
		set = func() { c.Ordered = &elem }
	case Named:
		set = func() { c.Named = &elem }
	case Destructured:
		set = func() { c.Destructured = &elem }
	default:
		panic("invalid type in the union")
	}
}

func (c *containedElem[Ordered, Named, Destructured]) Get() ContainedElem {
	switch {
	case c.Ordered != nil:
		return *c.Ordered
	case c.Named != nil:
		return *c.Named
	case c.Destructured != nil:
		return *c.Destructured
	default:
		return nil
	}
}

func (c *containedElem[Ordered, Named, Destructured]) Lookup() (ContainedElem, bool) {
	val := c.Get()

	return val, val != nil
}

func (c *containedElem[Ordered, Named, Destructured]) HasValue() bool {
	return c.Get() != nil
}

type noneContainedElem struct{}

func (noneContainedElem) GetExpr() Expr {
	panic("implement me")
}

func (noneContainedElem) IsNamed() bool {
	panic("implement me")
}

func (noneContainedElem) IsDeep() bool {
	panic("implement me")
}

func (noneContainedElem) IsDestructured() bool {
	panic("implement me")
}
