package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/ByronLiang/aws-gw-lambda/util"
)

func main() {
	sess, err := util.GetAwsSession()
	if err != nil {
		log.Println(err)
		return
	}
	s3Obj := s3.New(sess)
	nameSet, err := util.ListBuck(s3Obj)
	if err != nil {
		return
	}
	// 批量重压缩并上传
	for name := range nameSet {
		if !strings.HasPrefix(name, util.BackupBucketPrefix) {
			err = util.CreateBackupBucket(s3Obj, name)
			if err != nil {
				log.Println("create backup bucket error", err.Error())
				continue
			}
			objectList, err := util.RefreshObject(sess, name)
			if err != nil {
				log.Println("refresh error")
				continue
			}
			fmt.Println(objectList)
		}
	}
	//objectList, num := util.LoadBuck()
	//fmt.Println(objectList)
	//fmt.Println(len(objectList))
	//fmt.Println(num)
}
