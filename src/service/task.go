package service

import (
	"fmt"

	"github.com/spaco/affiliate/src/service/db"
	pg "github.com/spaco/affiliate/src/service/postgresql"
)

func SyncCryptocurrency(newCurrency []*db.CryptocurrencyInfo, updateRateCur []*db.CryptocurrencyInfo) {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	pg.AddBatchCryptocurrency(tx, newCurrency...)
	pg.UpdateBatchRate(tx, updateRateCur...)
	checkErr(tx.Commit())
	commit = true
}

const tellerReqName = "teller:req"

func GetTellerReq() int64 {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	intVal, _, _ := pg.GetKvStore(tx, tellerReqName)
	checkErr(tx.Commit())
	commit = true
	return intVal
}

func UpdateTellerReq(val int64) {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	pg.SaveKvStore(tx, tellerReqName, val, "")
	checkErr(tx.Commit())
	commit = true
}

func ProcessDeposit(batch []db.DepositRecord, req int64) {
	tx, commit := db.BeginTx()
	defer db.Rollback(tx, &commit)
	for _, dr := range batch {
		mapping, found := pg.QueryMappingDepositAddr(tx, dr.Address, dr.CurrencyType)
		if !found {
			panic(fmt.Sprintf("not found BuyAddrMapping for address:%s CurrencyType:%s", dr.Address, dr.CurrencyType))
		}
		dr.MappingId = mapping.Id
	}
	pg.SaveBatchDepositRecord(tx, batch...)
	// calculator reward
	commit = true
}
