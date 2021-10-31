package metadata

// M represents metadata of given host and parsed paths.
type M struct {
	Host string
	//				 paths	:   methods : metadata
	PathsMethods map[string]map[string]*MethodMetadata
}

// MethodMetadata describes method metadata.
type MethodMetadata struct {
	ContentType   string
	ParamsIn      string            // body / query
	ParamsType    string            // object / string / array
	ParamsValType map[string]string // name => type
}

// NewMethodMetadata describes path metadata.
func NewMethodMetadata() *MethodMetadata {
	return &MethodMetadata{
		ParamsValType: make(map[string]string),
	}
}

// New creates a new Metadata.
func New() *M {
	return &M{
		PathsMethods: make(map[string]map[string]*MethodMetadata),
	}
}

// AddMethod adds method to given path metadata.
func (m *M) AddMethod(path, method string) {
	pathMeta, ok := m.PathsMethods[path]
	if !ok {
		m.PathsMethods[path] = make(map[string]*MethodMetadata)
		pathMeta = m.PathsMethods[path]
	}
	pathMeta[method] = NewMethodMetadata()
}

// AddContentType adds content type to given path/method metadata.
func (m *M) AddContentType(path, method, ct string) {
	_, ok := m.PathsMethods[path]
	if !ok {
		m.PathsMethods[path] = make(map[string]*MethodMetadata)
	}

	methodMeta, ok := m.PathsMethods[path][method]
	if !ok {
		m.PathsMethods[path][method] = NewMethodMetadata()
		methodMeta = m.PathsMethods[path][method]
	}

	methodMeta.ContentType = ct
	m.PathsMethods[path][method] = methodMeta
}

// AddParamsIn adds location of parameters (query or object) to given path/method metadata.
func (m *M) AddParamsIn(path, method, pi string) {
	_, ok := m.PathsMethods[path]
	if !ok {
		m.PathsMethods[path] = make(map[string]*MethodMetadata)
	}

	methodMeta, ok := m.PathsMethods[path][method]
	if !ok {
		m.PathsMethods[path][method] = NewMethodMetadata()
		methodMeta = m.PathsMethods[path][method]
	}

	methodMeta.ParamsIn = pi
	m.PathsMethods[path][method] = methodMeta
}

// AddParamsType adds parameter type to given path/method metadata.
func (m *M) AddParamsType(path, method, pt string) {
	_, ok := m.PathsMethods[path]
	if !ok {
		m.PathsMethods[path] = make(map[string]*MethodMetadata)
	}

	methodMeta, ok := m.PathsMethods[path][method]
	if !ok {
		m.PathsMethods[path][method] = NewMethodMetadata()
		methodMeta = m.PathsMethods[path][method]
	}

	methodMeta.ParamsType = pt
	m.PathsMethods[path][method] = methodMeta
}

// AddParamsValType adds value and field type to given path/method metadata.
func (m *M) AddParamsValType(path, method, v, t string) {
	_, ok := m.PathsMethods[path]
	if !ok {
		m.PathsMethods[path] = make(map[string]*MethodMetadata)
	}

	methodMeta, ok := m.PathsMethods[path][method]
	if !ok {
		m.PathsMethods[path][method] = NewMethodMetadata()
		methodMeta = m.PathsMethods[path][method]
	}

	methodMeta.ParamsValType[v] = t
	m.PathsMethods[path][method] = methodMeta
}
