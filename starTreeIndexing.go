package main

import (
	"fmt"
)

type starNode struct {
	Name         string
	RecordCount  int
	Records      [][]string
	Aggregations map[string]int
	Parent       *starNode
	Children     []*starNode
}

func newStarNode(name string, recordCount int, records [][]string, aggregations map[string]int, parent *starNode, children []*starNode) *starNode {
	return &starNode{
		Name:         name,
		RecordCount:  recordCount,
		Records:      records,
		Aggregations: aggregations,
		Parent:       parent,
		Children:     children,
	}
}

func dropReturn(records [][]string, index int) ([]string, [][]string) {
	row := records[index]
	records = append(records[:index], records[index+1:]...)
	return row, records
}

func recursive(rootNode *starNode, i int, dimensions []string) {
	if rootNode.RecordCount > splitValue && i < len(dimensions) {
		uniqueValues := make(map[string]int)
		for _, record := range rootNode.Records {
			uniqueValues[record[i]]++
		}
		for k, v := range uniqueValues {
			if v > splitValue {
				var subRecords [][]string
				for _, record := range rootNode.Records {
					if record[i] == k {
						subRecords = append(subRecords, record)
					}
				}
				newNode := newStarNode(dimensions[i]+"_"+k, len(subRecords), subRecords, map[string]int{}, rootNode, nil)
				fmt.Println("Created New Node:", rootNode.Name, newNode.Name, newNode.RecordCount, "Dimension:", dimensions[i])
				recursive(newNode, i+1, dimensions)
				//tempRecordCount := len(rootNode.Records)
				for j := range rootNode.Records {
					if rootNode.Records[j][i] == k {
						_, rootNode.Records = dropReturn(rootNode.Records, j)
						rootNode.RecordCount = len(rootNode.Records)
					}
				}
			}
		}
		if i+1 < len(dimensions) {
			recursive(rootNode, i+1, dimensions)
		}
	}
}

const splitValue = 1000

//func main() {
//	// df := pd.read_csv
