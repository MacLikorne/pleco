package aws

import (
	"github.com/MacLikorne/pleco/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	log "github.com/sirupsen/logrus"
	"time"
)

type elasticacheCluster struct {
	ClusterIdentifier  string
	ReplicationGroupId string
	ClusterCreateTime  time.Time
	ClusterStatus      string
	TTL                int64
	IsProtected        bool
}

func ElasticacheSession(sess session.Session, region string) *elasticache.ElastiCache {
	return elasticache.New(&sess, &aws.Config{Region: aws.String(region)})
}

func listTaggedElasticacheDatabases(svc elasticache.ElastiCache, tagName string) ([]elasticacheCluster, error) {
	var taggedClusters []elasticacheCluster

	result, err := svc.DescribeCacheClusters(nil)
	if err != nil {
		return nil, err
	}

	if len(result.CacheClusters) == 0 {
		return nil, nil
	}

	for _, cluster := range result.CacheClusters {
		tags, err := svc.ListTagsForResource(
			&elasticache.ListTagsForResourceInput{
				ResourceName: aws.String(*cluster.ARN),
			},
		)
		if err != nil {
			if *cluster.CacheClusterStatus == "available" {
				log.Errorf("Can't get tags for Elasticache cluster: %s", *cluster.CacheClusterId)
			}
			continue
		}

		// required for replicas deletion
		replicationGroupId := ""
		if cluster.ReplicationGroupId != nil {
			replicationGroupId = *cluster.ReplicationGroupId
		}

		_, ttl, isProtected, _, _ := utils.GetEssentialTags(tags.TagList, tagName)

		taggedClusters = append(taggedClusters, elasticacheCluster{
			ClusterIdentifier:    *cluster.CacheClusterId,
			ReplicationGroupId:	  replicationGroupId,
			ClusterCreateTime:    *cluster.CacheClusterCreateTime,
			ClusterStatus:        *cluster.CacheClusterStatus,
			TTL:                  ttl,
			IsProtected: isProtected,
		})

	}

	return taggedClusters, nil
}

func deleteElasticacheCluster(svc elasticache.ElastiCache, cluster elasticacheCluster) error {
	if cluster.ClusterStatus == "deleting" {
		log.Infof("Elasticache cluster %s is already in deletion process, skipping...", cluster.ClusterIdentifier)
		return nil
	} else {
		log.Infof("Deleting Elasticache cluster %s in %s, expired after %d seconds",
			cluster.ClusterIdentifier, *svc.Config.Region, cluster.TTL)
	}

	// with replicas
	if cluster.ReplicationGroupId != "" {
		_, err := svc.DeleteReplicationGroup(
			&elasticache.DeleteReplicationGroupInput{
				ReplicationGroupId:      aws.String(cluster.ReplicationGroupId),
				RetainPrimaryCluster:    aws.Bool(false),
			},
		)
		if err != nil {
			return err
		}
	}

	_, err := svc.DeleteCacheCluster(
		&elasticache.DeleteCacheClusterInput{
			CacheClusterId: aws.String(cluster.ClusterIdentifier),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func DeleteExpiredElasticacheDatabases(svc elasticache.ElastiCache, tagName string, dryRun bool) {
	clusters, err := listTaggedElasticacheDatabases(svc, tagName)
	region := svc.Config.Region
	if err != nil {
		log.Errorf("can't list Elasticache databases: %s\n", err)
		return
	}

	var expiredClusters []elasticacheCluster
	for _, cluster := range clusters {
		if utils.CheckIfExpired(cluster.ClusterCreateTime, cluster.TTL, "elasticache: " + cluster.ClusterIdentifier) && !cluster.IsProtected{
			expiredClusters = append(expiredClusters, cluster)
		}
	}

	count, start:= utils.ElemToDeleteFormattedInfos("expired Elasticache database", len(expiredClusters), *region)

	log.Debug(count)

	if dryRun || len(expiredClusters) == 0 {
		return
	}

	log.Debug(start)

	for _, cluster := range clusters {
		deletionErr := deleteElasticacheCluster(svc, cluster)
		if deletionErr != nil {
			log.Errorf("Deletion Elasticache cluster error %s/%s: %s",
					cluster.ClusterIdentifier, *svc.Config.Region, err)
			}
	}

}