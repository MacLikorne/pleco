module github.com/MacLikorne/pleco/providers/k8s

go 1.16

replace github.com/MacLikorne/pleco/utils => ../../utils

require (
	github.com/MacLikorne/pleco/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
)
