package service

import (
	"github.com/spaco/affiliate/src/service/db"
	pg "github.com/spaco/affiliate/src/service/postgresql"
	client "github.com/spaco/affiliate/src/teller_client"
)

func AllCryptocurrencyMap() map[string]db.CryptocurrencyInfo {
	slice := AllCryptocurrency()
	m := make(map[string]db.CryptocurrencyInfo, len(slice))
	for _, info := range slice {
		m[info.ShortName] = info
	}
	return m
}
func GetCryptocurrency(currencyType string) *db.CryptocurrencyInfo {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	info := pg.GetCryptocurrency(tx, currencyType)
	checkErr(tx.Commit())
	commit = true
	return info
}
func AllCryptocurrency() []db.CryptocurrencyInfo {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	all := pg.AllCryptocurrency(tx)
	checkErr(tx.Commit())
	commit = true
	return all
}

func AddBatchCryptocurrency(batch []db.CryptocurrencyInfo) {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	pg.AddBatchCryptocurrency(tx, batch)
	checkErr(tx.Commit())
	commit = true
}

func MappingDepositAddr(address string, currencyType string, ref string) (string, error) {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	buyAddrMapping := pg.QueryMappingDepositAddr(tx, address, currencyType)
	if buyAddrMapping != nil {
		checkErr(tx.Commit())
		commit = true
		return buyAddrMapping.DepositAddr, nil
	}
	depositAddr, err := client.Bind(currencyType, address)
	if err != nil {
		return "", err
	}
	pg.SaveDepositAddrMapping(tx, address, currencyType, ref, depositAddr)
	checkErr(tx.Commit())
	commit = true
	return depositAddr, nil
}

func CheckMappingAddr(address string, currencyType string) bool {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	mapping := pg.QueryMappingDepositAddr(tx, address, currencyType)
	checkErr(tx.Commit())
	commit = true
	return mapping != nil
}

func QueryDepositRecord(address string, currencyType string) []db.DepositRecord {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	res := pg.QueryDepositRecord(tx, address, currencyType)
	checkErr(tx.Commit())
	commit = true
	return res
}

func SyncCryptocurrency(slice []db.CryptocurrencyInfo) {
	currencyMap := AllCryptocurrencyMap()
	newCurrency := make([]db.CryptocurrencyInfo, 0, 4)
	updateRateCur := make([]db.CryptocurrencyInfo, 0, 4)
	for _, info := range slice {
		if old, ok := currencyMap[info.ShortName]; ok {
			if old.Rate != info.Rate {
				updateRateCur = append(updateRateCur, info)
			}
		} else {
			newCurrency = append(newCurrency, info)
		}
	}
	syncCryptocurrency(newCurrency, updateRateCur)
}
