package vmdb

const (
	clusterQueryPath    = "/select/0/prometheus/api/v1/query"
	singleNodeQueryPath = "/api/v1/query"
)

type URLProvider interface {
	Get() string
	Post() string
}

type SingleNodeURL struct {
	VMURL string
}

func (s *SingleNodeURL) Get() string {
	return s.VMURL + singleNodeQueryPath
}

func (s *SingleNodeURL) Post() string {
	return s.VMURL
}

type ClusterURL struct {
	VMAgentURL  string
	VMSelectURL string
}

func (s *ClusterURL) Get() string {
	return s.VMSelectURL + clusterQueryPath
}

func (s *ClusterURL) Post() string {
	return s.VMAgentURL
}
