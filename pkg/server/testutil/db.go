package testutil

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/onsi/ginkgo"

	"github.com/armsnyder/othelgo/pkg/server"
)

// testTableName returns a table name that is unique for the ginkgo test node, allowing tests to
// run in parallel using different tables.
func testTableName() string {
	return fmt.Sprintf("Othelgo-%d", ginkgo.GinkgoParallelNode())
}

// clearOthelgoTable deletes and recreates the othelgo dynamodb table.
func clearOthelgoTable() {
	db := server.LocalDB()
	tableName := testTableName()

	_, _ = db.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})

	err := server.EnsureTable(context.Background(), db, tableName)
	if err != nil {
		panic(fmt.Errorf("testutil: Failed to clear dynamodb table: %w", err))
	}
}

func dumpTable() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	output, err := server.LocalDB().ScanWithContext(ctx, &dynamodb.ScanInput{
		TableName: aws.String(testTableName()),
	})

	if err != nil {
		log.Printf("testutil: Error scanning DB: %v", err)
		return
	}

	if len(output.Items) == 0 {
		log.Println("testutil: No items in DB")
		return
	}

	log.Println("testutil: DB table dump:")

	for i, item := range output.Items {
		var data map[string]interface{}
		if err := dynamodbattribute.UnmarshalMap(item, &data); err != nil {
			log.Printf("testutil: Error unmarshalling item #%d: %v", i, err)
			continue
		}

		log.Printf("testutil: item #%d: %v", i, data)
	}
}
