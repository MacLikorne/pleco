module github.com/MacLikorne/pleco/core

go 1.16

replace github.com/MacLikorne/pleco/providers/aws => ../providers/aws

replace github.com/MacLikorne/pleco/providers/k8s => ../providers/k8s

require (
	github.com/Qovery/pleco v0.7.19
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
)
