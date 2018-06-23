package main

import (
    "context"
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
	"io/ioutil"
	"os"
	"os/user"
	"path"
    "net/http"
    "bufio"
    "time"
    "encoding/json"
)

func pay_invoice(client lnrpc.LightningClient, payment_request string) string {
	ctx := context.Background()
	sendRequestResp, err := client.SendPaymentSync(ctx, &lnrpc.SendRequest{PaymentRequest: payment_request})

	if err != nil {
		fmt.Println("Cannot send payment from node:", err)
		return err.Error()
	}
	return sendRequestResp.String()
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJson(url string, target *Foo) error {
    r, err := myClient.Get(url)
    if err != nil {
        return err
    }

    //b, err := ioutil.ReadAll(r.Body)
    defer r.Body.Close()

    //fmt.Println(string(b))
    return json.NewDecoder(r.Body).Decode(target)
}

type DataStruct struct {
    Invoice         string
    Buttonpressed   string
    AmountInSatoshi int64
}

type Foo struct {
    Data    DataStruct
    Message string
    Success bool
}



func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Cannot get current user:", err)
		return
	}
	tlsCertPath := path.Join(usr.HomeDir, ".lnd/tls.cert")
	macaroonPath := path.Join(usr.HomeDir, ".lnd/admin.macaroon")

	tlsCreds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		fmt.Println("Cannot get node tls credentials", err)
		return
	}

	macaroonBytes, err := ioutil.ReadFile(macaroonPath)
	if err != nil {
		fmt.Println("Cannot read macaroon file", err)
		return
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		fmt.Println("Cannot unmarshal macaroon", err)
		return
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(macaroons.NewMacaroonCredential(mac)),
	}

	fmt.Print("Trying to connect to lnd...")
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:10009"), opts...)
	if err != nil {
		fmt.Println("cannot dial to lnd", err)
		return
	}
	client := lnrpc.NewLightningClient(conn)
	fmt.Println("ok")

    for {
        foo1 := Foo{} // or &Foo{}
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Enter button: ")
        text, _ := reader.ReadString('\n')
        fmt.Println(text)
        switch text {
        case "w\n":
            getJson("http://lnplays.com/getInvoice/up", &foo1)
        case "a\n":
            getJson("http://lnplays.com/getInvoice/left", &foo1)
        case "s\n":
            getJson("http://lnplays.com/getInvoice/down", &foo1)
        case "d\n":
            getJson("http://lnplays.com/getInvoice/right", &foo1)
        case " \n":
            getJson("http://lnplays.com/getInvoice/a", &foo1)
        case "b\n":
            getJson("http://lnplays.com/getInvoice/b", &foo1)
        default:
            fmt.Println("unknown key")
            continue
        }

        fmt.Println(foo1.Data.Invoice)
        fmt.Println(pay_invoice(client, foo1.Data.Invoice))
    }
}
