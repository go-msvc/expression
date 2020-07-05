package expression

//IContextr context in which expressions are evaluated
type IContext interface {
	Set(name string, value interface{})
	Get(name string) interface{}
	//Identifier(s string) IIdentifier
}

func NewContext() IContext {
	return &context{
		data: map[string]interface{}{},
	}
}

type context struct {
	data map[string]interface{}
}

func (ctx context) Identifier(s string) IIdentifier {
	return nil
}

func (ctx *context) Set(name string, value interface{}) {
	ctx.data[name] = value
}
func (ctx context) Get(name string) interface{} {
	v, ok := ctx.data[name]
	if !ok {
		return nil
	}
	return v
}
