package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	tigerbeetle_go "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	// "github.com/tigerbeetle/tigerbeetle-go"
	// "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

var (
	// só pra fazer o go parar de me encher os patavá
	_ = spew.Config
	_ = time.Sleep
)

var (
	httpAddr        string
	tigerbeetleHost string
)

const (
	TIGERBETTLE_MAX_CONCURRENCY = 256
)

func init() {
	portFromEnv := os.Getenv("PORT")
	if portFromEnv == "" {
		portFromEnv = "3001"
	}
	flag.StringVar(&httpAddr, "addr", fmt.Sprintf(":%s", portFromEnv), "Address where to listen for the server")

	flag.StringVar(&tigerbeetleHost, "t", os.Getenv("TB_ADDRESS"), "How to connect to tigerbeetle")
	flag.Parse()
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("can't prepare server: %s", err)
	}
	defer app.Close()
	err = app.Setup()
	if err != nil {
		log.Fatalf("can't setup app: %s", err)
	}
	log.Printf("Starting http server on %s", httpAddr)
	log.Printf("Pau na máquina!")
	err = http.ListenAndServe(httpAddr, app)
	if err != nil {
		log.Fatalf("can't start http server: %s", err)
	}
}

type App struct {
	tigerbeetle tigerbeetle_go.Client
}

func NewApp() (*App, error) {
	client, err := tigerbeetle_go.NewClient(types.ToUint128(0), strings.Split(tigerbeetleHost, ","), TIGERBETTLE_MAX_CONCURRENCY)
	if err != nil {
		return nil, err
	}
	return &App{tigerbeetle: client}, nil
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "foi!")
}

// pra dar crédito/limite pras contas precisa de uma conta de funding
var TIGERBEETLE_FUNDING_ACCOUNT_ID = uint64(9999999)

// contas de tal ledger só conseguem transacionar com tal ledger
// basicamente id de moeda
const TIGERBEETLE_DEFAULT_LEDGER = 1

// tipo de conta, tipo poupança, CC, caixa dois, sla
const TIGERBEETLE_DEFAULT_CODE = 1

var TIGERBEETLE_USER_ACCOUNTS_FLAGS = (types.AccountFlags{
	DebitsMustNotExceedCredits: true,
}).ToUint16() | (1 << 3) // 3 é a flag pra salvar o extrato

func (a *App) Setup() error {
	var limites = map[uint64]uint64{
		1: 1000 * 100,
		2: 800 * 100,
		3: 10000 * 100,
		4: 100000 * 100,
		5: 5000 * 100,
	}
	contas := make([]types.Account, len(limites)+1)
	transferencias := make([]types.Transfer, len(limites))
	for i, limite := range limites {
		contas[i-1] = types.Account{
			ID:     types.ToUint128(i),
			Ledger: TIGERBEETLE_DEFAULT_LEDGER,
			Code:   TIGERBEETLE_DEFAULT_CODE,
			Flags:  TIGERBEETLE_USER_ACCOUNTS_FLAGS,
		}
		transferencias[i-1] = types.Transfer{
			ID:              types.ToUint128(i),
			DebitAccountID:  types.ToUint128(TIGERBEETLE_FUNDING_ACCOUNT_ID),
			CreditAccountID: types.ToUint128(i),
			Amount:          types.ToUint128(limite),
		}
	}
	contas[len(limites)] = types.Account{
		ID:     types.ToUint128(TIGERBEETLE_FUNDING_ACCOUNT_ID),
		Ledger: TIGERBEETLE_DEFAULT_LEDGER,
		Code:   TIGERBEETLE_DEFAULT_CODE,
		Flags:  0, // sem restrições
	}

	_, err := a.tigerbeetle.CreateAccounts(contas)
	if err != nil {
		return err
	}
	_, err = a.tigerbeetle.CreateTransfers(transferencias)
	return err
}

func (a *App) Transfer(from uint64, to uint64, amount int, id uint64) error {
	if id == 0 {
		id = uint64(time.Now().UnixMicro())
	}
	/*transfers*/ _, err := a.tigerbeetle.CreateTransfers([]types.Transfer{
		{
			ID:              types.ToUint128(id),
			DebitAccountID:  types.ToUint128(from),
			CreditAccountID: types.ToUint128(to),
			Amount:          types.ToUint128(uint64(amount)),
			Ledger:          TIGERBEETLE_DEFAULT_LEDGER,
			Code:            TIGERBEETLE_DEFAULT_CODE,
		},
	})
	// spew.Dump(transfers)
	return err
}

func (a *App) Close() error {
	a.tigerbeetle.Close()
	return nil
}
