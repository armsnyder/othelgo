package testutil

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"
	"unicode/utf8"

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

// dumpTable scans the full table and prints it to the log output.
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
		var itemFields map[string]interface{}
		if err := dynamodbattribute.UnmarshalMap(item, &itemFields); err != nil {
			log.Printf("testutil: Error unmarshalling item #%d: %v", i, err)
			continue
		}

		// Fix encoding of any []byte fields so that they are more readable.
		for key, valueAny := range itemFields {
			if valueBytes, ok := valueAny.([]byte); ok {
				if utf8.Valid(valueBytes) {
					itemFields[key] = string(valueBytes)
				} else {
					itemFields[key] = base64.StdEncoding.EncodeToString(valueBytes)
				}
			}
		}

		log.Printf("testutil: item #%d: %v", i, itemFields)
	}
}
