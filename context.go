package expression

//IContextr context in which expressions are evaluated
type IContext interface {
	Identifier(s string) IIdentifier
}

func NewContext() IContext {
	return &context{}
}

type context struct{}

func (ctx context) Identifier(s string) IIdentifier {
	return nil
}
