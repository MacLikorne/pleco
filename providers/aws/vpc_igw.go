package aws

import (
	"github.com/MacLikorne/pleco/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type InternetGateway struct {
	Id           string
	CreationDate time.Time
	ttl          int64
	IsProtected  bool
}

func getInternetGatewaysByVpcId (ec2Session ec2.EC2, vpcId string) []*ec2.InternetGateway{
	input := &ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("attachment.vpc-id"),
				Values: []*string{aws.String(vpcId)},
			},
		},
	}

	gateways, err := ec2Session.DescribeInternetGateways(input)
	if err != nil {
		log.Error(err)
	}

	return gateways.InternetGateways
}

func SetInternetGatewaysIdsByVpcId (ec2Session ec2.EC2, vpc *VpcInfo, waitGroup *sync.WaitGroup, tagName string) {
	defer waitGroup.Done()
	var internetGateways []InternetGateway

	gateways := getInternetGatewaysByVpcId(ec2Session, *vpc.VpcId)

	for _, gateway := range gateways {
		creationDate, ttl, isProtected, _, _ := utils.GetEssentialTags(gateway.Tags,tagName)

		var gatewayStruct = InternetGateway{
			Id: *gateway.InternetGatewayId,
			CreationDate: creationDate,
			ttl: ttl,
			IsProtected: isProtected,
		}

		internetGateways = append(internetGateways, gatewayStruct)
	}

	vpc.InternetGateways= internetGateways
}

func DeleteInternetGatewaysByIds (ec2Session ec2.EC2, internetGateways []InternetGateway) {
	for _, internetGateway := range internetGateways {
		if utils.CheckIfExpired(internetGateway.CreationDate, internetGateway.ttl, "vpc internet gateway: " + internetGateway.Id) && !internetGateway.IsProtected {
			_, err := ec2Session.DeleteInternetGateway(
				&ec2.DeleteInternetGatewayInput{
					InternetGatewayId: aws.String(internetGateway.Id),
				},
			)

			if err != nil {
				log.Error(err)
			}
		}
	}
}
