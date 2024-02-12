package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
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

type SubmitTransactionRequest struct {
	Valor     uint64 `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

type SubmitTransactionResponse struct {
	Limite uint64 `json:"limite"`
	Saldo  int64  `json:"saldo"` // pode ser negativo
}

type ExtratoReponse struct {
	Saldo             ExtratoSaldoResponse `json:"saldo"`
	UltimasTransacoes []ExtratoTransacao   `json:"ultimas_transacoes"`
}

type ExtratoSaldoResponse struct {
	Total            int64     `json:"total"`
	TimestampExtrato time.Time `json:"data_extrato"`
	Limite           uint64    `json:"limite"`
}

type ExtratoTransacao struct {
	Valor              uint64    `json:"valor"`
	Tipo               string    `json:"tipo"`
	Descricao          string    `json:"descricao"`
	TimestampTransacao time.Time `json:"realizada_em"`
}

func (a *App) GetSaldo(cliente int) (SubmitTransactionResponse, error) {
	var response SubmitTransactionResponse
	account, err := a.tigerbeetle.LookupAccounts([]types.Uint128{
		types.ToUint128(uint64(cliente)),
	})
	if err != nil {
		return response, err
	}
	response.Limite = account[0].UserData64
	response.Saldo = 0
	saldoParcial := big.NewInt(0)

	stepInt := &big.Int{}
	stepInt.SetBytes(account[0].CreditsPending[:])
	saldoParcial.Add(saldoParcial, stepInt)

	stepInt.SetBytes(account[0].CreditsPosted[:])
	saldoParcial.Add(saldoParcial, stepInt)

	stepInt.SetBytes(account[0].DebitsPending[:])
	saldoParcial.Sub(saldoParcial, stepInt)

	stepInt.SetBytes(account[0].DebitsPosted[:])
	saldoParcial.Sub(saldoParcial, stepInt)

	response.Saldo = int64(saldoParcial.Uint64()) - int64(response.Limite)

	return response, nil
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	spew.Dump(urlParts)
	if len(urlParts) > 1 && urlParts[0] == "" {
		urlParts = urlParts[1:]
	}
	if len(urlParts) < 3 {
		// todas as rotas são /clientes/:id/valorchumbado
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if urlParts[0] != "clientes" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var err error
	clienteId, err := strconv.Atoi(urlParts[1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w)
	if r.Method == http.MethodPost && urlParts[2] == "transacoes" {
		var request SubmitTransactionRequest
		err := decoder.Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(request.Descricao) > 10 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch request.Tipo {
		case "d":
			err = a.Transfer(
				uint64(clienteId),
				TIGERBEETLE_FUNDING_ACCOUNT_ID,
				request.Valor,
				0,
				request.Descricao,
			)
		case "c":
			err = a.Transfer(
				TIGERBEETLE_FUNDING_ACCOUNT_ID,
				uint64(clienteId),
				request.Valor,
				0,
				request.Descricao,
			)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		response, err := a.GetSaldo(clienteId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		encoder.Encode(response)
		return
	}
	accountID := types.ToUint128(uint64(clienteId))
	if r.Method == http.MethodGet && urlParts[2] == "extrato" {
		filter := types.AccountFilter{
			AccountID: accountID,
			Limit:     10,
			Flags: types.AccountFilterFlags{
				Debits:   true,
				Credits:  true,
				Reversed: false,
			}.ToUint32(),
		}
		saldo, err := a.GetSaldo(clienteId)
		ret := ExtratoReponse{
			Saldo: ExtratoSaldoResponse{
				Total:            saldo.Saldo,
				Limite:           saldo.Limite,
				TimestampExtrato: time.Now(),
			},
			UltimasTransacoes: make([]ExtratoTransacao, 0, 10),
		}

		transfers_filtered, err := a.tigerbeetle.GetAccountTransfers(filter)
		log.Printf("%s", err)
		for i, transfer := range transfers_filtered {
			tipo := "d"
			descricaoBytes := transfer.UserData128.Bytes()
			descricaoSize := descricaoBytes[0]
			descricao := ""
			if descricaoSize < 16 && descricaoSize > 0 {
				descricao = string(descricaoBytes[1:descricaoSize])
			}
			if transfer.CreditAccountID.String() == accountID.String() {
				tipo = "c"
			}
			ret.UltimasTransacoes = append(ret.UltimasTransacoes, ExtratoTransacao{
				Valor:              transfer.Amount.BigInt().Uint64(),
				Tipo:               tipo,
				Descricao:          descricao,
				TimestampTransacao: time.UnixMicro(int64(transfer.Timestamp) / 1000),
			})
			spew.Dump(transfer, i)

		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = encoder.Encode(ret)
		if err != nil {
			log.Printf("%s", err)
		}
		// log.Printf("%s", err)
		// var ret ExtratoRespons
		history, err := a.tigerbeetle.GetAccountHistory(filter)
		log.Printf("%s", err)
		spew.Dump(history)
		return
		// spew.Dump(transfers_filtered)

	}
	w.WriteHeader(http.StatusNotFound)
	// fmt.Fprintf(w, "foi!")
}

// pra dar crédito/limite pras contas precisa de uma conta de funding
var TIGERBEETLE_FUNDING_ACCOUNT_ID = uint64(9999999)

// contas de tal ledger só conseguem transacionar com tal ledger
// basicamente id de moeda
const TIGERBEETLE_DEFAULT_LEDGER = 1

// tipo de conta, tipo poupança, CC, caixa dois, sla
const TIGERBEETLE_DEFAULT_CODE = 1

var TIGERBEETLE_USER_ACCOUNTS_FLAGS = types.AccountFlags{
	DebitsMustNotExceedCredits: true,
	History:                    true,
}.ToUint16()

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
			ID:         types.ToUint128(i),
			Ledger:     TIGERBEETLE_DEFAULT_LEDGER,
			Code:       TIGERBEETLE_DEFAULT_CODE,
			UserData64: limite,
			Flags:      TIGERBEETLE_USER_ACCOUNTS_FLAGS,
		}
		transferencias[i-1] = types.Transfer{
			ID:              types.ToUint128(i),
			DebitAccountID:  types.ToUint128(TIGERBEETLE_FUNDING_ACCOUNT_ID),
			CreditAccountID: types.ToUint128(i),
			Amount:          types.ToUint128(limite),
			Ledger:          TIGERBEETLE_DEFAULT_LEDGER,
			Code:            TIGERBEETLE_DEFAULT_CODE,
		}
	}
	contas[len(limites)] = types.Account{
		ID:     types.ToUint128(TIGERBEETLE_FUNDING_ACCOUNT_ID),
		Ledger: TIGERBEETLE_DEFAULT_LEDGER,
		Code:   TIGERBEETLE_DEFAULT_CODE,
		Flags:  0, // sem restrições
	}

	accountsResult, err := a.tigerbeetle.CreateAccounts(contas)
	spew.Dump(accountsResult)
	if err != nil {
		return err
	}
	transferenciasResult, err := a.tigerbeetle.CreateTransfers(transferencias)
	spew.Dump(transferenciasResult)
	return err
}

func (a *App) Transfer(from uint64, to uint64, amount uint64, id uint64, description string) error {
	if id == 0 {
		id = uint64(time.Now().UnixMicro())
	}
	// TODO: como caraglios codificar descrição?

	/*transfers*/
	descsize := 16
	descbytes := make([]byte, descsize)

	descriptionAsBytes := []byte(description)
	descriptionSize := len(descriptionAsBytes)
	if descriptionSize > 15 {
		descriptionSize = 15
	}
	descbytes[0] = byte(descriptionSize)
	for i := 0; i < descriptionSize; i++ {
		descbytes[i+1] = descriptionAsBytes[i]
	}
	_, err := a.tigerbeetle.CreateTransfers([]types.Transfer{
		{
			ID:              types.ToUint128(id),
			DebitAccountID:  types.ToUint128(from),
			CreditAccountID: types.ToUint128(to),
			Amount:          types.ToUint128(uint64(amount)),
			Ledger:          TIGERBEETLE_DEFAULT_LEDGER,
			Code:            TIGERBEETLE_DEFAULT_CODE,
			UserData128:     types.Uint128(descbytes),
		},
	})
	// spew.Dump(transfers)
	return err
}

func (a *App) Close() error {
	a.tigerbeetle.Close()
	return nil
}
