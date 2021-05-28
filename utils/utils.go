package utils

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type Tag struct {
	_     struct{} `type:"structure"`
	Key   *string  `type:"string"`
	Value *string  `type:"string"`
}

type MyTag struct {
	_     struct{} `type:"structure"`
	Key   string  `type:"string"`
	Value string  `type:"string"`
}

func GetEssentialTags(tagsInput interface{}, tagName string) (time.Time, int64, bool, string, string) {
	var creationDate = time.Time{}
	var ttl int64
	var isProtected bool
	var clusterId string
	var tag string
	var tags []MyTag

	switch typedTags := tagsInput.(type) {
		case []*rds.Tag:
			for _, elem := range typedTags {
				tags = append(tags, MyTag{Key: *elem.Key, Value: *elem.Value})
			}
		case []*ec2.Tag:
			for _, elem := range typedTags {
				tags = append(tags, MyTag{Key: *elem.Key, Value: *elem.Value})
			}
		case []*iam.Tag:
			for _, elem := range typedTags {
				tags = append(tags, MyTag{Key: *elem.Key, Value: *elem.Value})
			}
		case []*kms.Tag:
			for _, elem := range typedTags {
				tags = append(tags, MyTag{Key: *elem.TagKey, Value: *elem.TagValue})
			}
		case []*s3.Tag:
			for _, elem := range typedTags {
				tags = append(tags, MyTag{Key: *elem.Key, Value: *elem.Value})
			}
		case []*elbv2.Tag:
			for _, elem := range typedTags {
				tags = append(tags, MyTag{Key: *elem.Key, Value: *elem.Value})
			}
		case []*elasticache.Tag:
			for _, elem := range typedTags {
				tags = append(tags, MyTag{Key: *elem.Key, Value: *elem.Value})
			}
		case []*Tag:
			for _, elem := range typedTags {
				tags = append(tags, MyTag{Key: *elem.Key, Value: *elem.Value})
			}
		case map[string]*string:
			for key, value := range typedTags {
				tags = append(tags, MyTag{Key: key, Value: *value})
			}
		default:
			log.Debugf("Can't parse tags %s.", tagsInput)
	}

	for i := range tags {
		switch tags[i].Key {
			case "creationDate":
				creationTime, _ := strconv.ParseInt(tags[i].Value, 10, 64)
				creationDate = time.Unix(creationTime,0)
			case "ttl":
				result, _ := strconv.ParseInt(tags[i].Value, 10, 64)
				ttl = result
			case "do_not_delete":
				result, _ := strconv.ParseBool(tags[i].Value)
				isProtected = result
			case "ClusterId":
				clusterId = tags[i].Value
			case tagName:
				tag = tags[i].Value
			default:
				continue
			}
	}

	return creationDate, ttl, isProtected, clusterId, tag
}

func CheckIfExpired(creationTime time.Time, ttl int64, resourceName string) bool {
	expirationTime := creationTime.Add(time.Duration(ttl) * time.Second)
	if ttl == 0 {
		return false
	}

	if creationTime.Year() < 1972 {
		log.Errorf("Creation date tag is missing. Can't check if resource %s is expired.", resourceName)
		return false
	}

	return time.Now().After(expirationTime)
}

func ElemToDeleteFormattedInfos(elemName string, arraySize int, region string) (string,string) {
	count := fmt.Sprintf("There is no %s to delete in region %s.", elemName,region)
	if arraySize == 1 {
		count = fmt.Sprintf("There is no %s to delete in region %s.", elemName,region)
	}
	if arraySize > 1 {
		count = fmt.Sprintf("There are %d %ss to delete in region %s.", arraySize, elemName,region)
	}

	start := fmt.Sprintf("Starting %s deletion for region %s.", elemName,region)


	return count, start
}


func AwsStringChecker(elem interface{ String() string }) string {
	if elem != nil {
		return elem.String()
	} else {
		return ""
	}
}

func getSlicedArray(arrayToSlice []*string, sliceRange int) [][]*string {
	var slicedArray [][]*string
	slicesCount := len(arrayToSlice) / sliceRange + 1

	if len(arrayToSlice) <= sliceRange {
		slicedArray = append(slicedArray, arrayToSlice)
	} else {
		for i := 0; i < slicesCount; i++ {
			if (i+1) * sliceRange > len(arrayToSlice) {
				slicedArray = append(slicedArray, arrayToSlice[i*sliceRange:len(arrayToSlice)-1])
			} else {
				slicedArray = append(slicedArray, arrayToSlice[i*sliceRange:(i+1)*sliceRange])
			}
		}
	}


	return slicedArray
}




