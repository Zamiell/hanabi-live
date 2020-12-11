package tables

// addUserTable adds a table ID to the list for the respective user
func addUserTable(userID int, tableID uint64, userMap map[int][]uint64) {
	tableList, ok := userMap[userID]
	if !ok {
		// This will be the first table for this user
		tableList = make([]uint64, 0)
	}
	tableList = append(tableList, tableID)
	userMap[userID] = tableList
}

// deleteUserTable deletes a table ID from the list for the respective user
func deleteUserTable(userID int, tableID uint64, userMap map[int][]uint64) {
	tableList, ok := userMap[userID]
	if !ok {
		// This user is not in the map,
		// which subsequently means that they are not present at any tables,
		// so this delete operation is a no-op
		return
	}

	i := indexOf(tableID, tableList)
	if i == -1 {
		// This table ID is not in the list for this user, so this delete operation is a no-op
		return
	}

	tableList = append(tableList[:i], tableList[i+1:]...)
	if len(tableList) == 0 {
		// This user is not present at any tables, so delete the entry for this user in the map
		delete(userMap, userID)
	} else {
		// Save the new table list to the map
		userMap[userID] = tableList
	}
}

// getUserTables gets the list of tables for the respective user
func getUserTables(userID int, userMap map[int][]uint64) []uint64 {
	if tablesList, ok := userMap[userID]; ok {
		return tablesList
	}
	return make([]uint64, 0)
}

func indexOf(value uint64, slice []uint64) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}

	return -1
}
