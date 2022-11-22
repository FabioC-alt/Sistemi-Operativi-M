package main

import (
	"fmt"
	"time"
)

const MAXPROC = 100 //massimo numero di processi
const NBT = 5       //numero di biciclette tradizionali
const NEB = 1       //numero di e-bike

const BT = 0   // tipo 0 -> richiesta bici tradizionale
const EB = 1   // tipo 1 -> richiesta e-bike
const FLEX = 2 // tipo 2 -> richiesta flessibile

const DIMBUF = 300

// nella richiesta devo passare sia l'id del client che il tipo di richiesta
type req struct {
	id   int // identificatore client
	tipo int // tipo di richiesta BT EB o FLEX
}

var richiestaBT = make(chan req, DIMBUF)
var richiestaEB = make(chan req, DIMBUF)
var richiestaFLEX = make(chan req, DIMBUF)
var rilascio = make(chan int)
var risorsa [MAXPROC]chan int // ogni goroutine avrà il suo canale privato

// per mettere su un protocollo di sincronizzazione
var done = make(chan int)
var termina = make(chan int)

func when(b bool, c chan req) chan req {
	if !b {
		return nil
	}
	return c
}

func client(r req) {
	var b int         //BT o EB
	if r.tipo == BT { //richiesta BT
		richiestaBT <- r
		fmt.Printf("[client %d] richiesta BICI TRADIZIONALE\n", r.id)
	} else if r.tipo == EB { //richiesta EB
		richiestaEB <- r
		fmt.Printf("[client %d] richiesta E-BIKE\n", r.id)
	} else { //richiesta FLEX
		richiestaFLEX <- r
		fmt.Printf("[client %d] richiesta FLEX\n", r.id)
	}
	//attende la risorsa i-esima che avrà valore BT o EB per indicare il tipo di risorsa utilizzata
	b = <-risorsa[r.id]
	if b == BT {
		fmt.Printf("[client %d] ottenuta BICI TRADIZIONALE\n", r.id)
	} else {
		fmt.Printf("[client %d] ottenuta EBIKE\n", r.id)
	}

	//tempo di uso della risorsa
	time.Sleep(time.Second * 2)

	//rilascio della risorsa
	rilascio <- b
	done <- r.id //comunico al main la terminazione
}

func server() { //ha lo scopo di gestire le risorse del pool
	var dispBT int = NBT
	var dispEB int = NEB
	var b int
	var r req

	//comando con guardia ripetitivo = un ciclo con comandi di guardia alternativi -> for con select
	for {
		time.Sleep(time.Second * 1)
		select {
		case b = <-rilascio:
			if b == EB { //restituzione EB
				dispEB++
				fmt.Println("[server] restituita EBIKE")
			} else { //restituzione BT
				dispBT++
				fmt.Println("[server] restituita BICI TRADIZIONALE")
			}
		case r = <-when(dispBT > 0, richiestaBT):
			dispBT--
			b = BT
			risorsa[r.id] <- b
			fmt.Printf("[server] allocata BT a cliente %d\n", r.id)
		case r = <-when(dispEB > 0, richiestaEB):
			dispEB--
			b = EB
			risorsa[r.id] <- b
			fmt.Printf("[server] allocata EB a cliente %d\n", r.id)
		case r = <-when(dispEB > 0, richiestaFLEX):
			dispEB--
			b = EB
			risorsa[r.id] <- b
			fmt.Printf("[server] allocata EB a cliente %d\n", r.id)
		case r = <-when(dispEB == 0 && dispBT > 0, richiestaFLEX):
			dispBT--
			b = BT
			risorsa[r.id] <- b
			fmt.Printf("[server] allocata BT a cliente %d\n", r.id)
		case r = <-when(dispEB == 0 && dispBT == 0, richiestaFLEX):
			richiestaEB <- r
			fmt.Printf("[server] cliente %d in ATTESA di un EB\n", r.id)

		case <-termina: // quando tutti i processi clienti hanno finito
			fmt.Println("FINE!")
			done <- 1
			return

		}
	}
}

func main() {
	var cli int
	var r req

	fmt.Printf("\n quanti clienti (max %d)? ", MAXPROC)
	fmt.Scanf("%d", &cli)
	fmt.Println("clienti:", cli)

	//inizializzazione array di canali
	for i := 0; i < cli; i++ {
		risorsa[i] = make(chan int)
	}

	//meglio creare prima il server e poi i client
	go server()

	for i := 0; i < cli; i++ {
		r.id = i
		r.tipo = i % 3
		go client(r)
	}

	//attesa della terminazione dei clienti:
	for i := 0; i < cli; i++ {
		<-done
	}
	termina <- 1 //terminazione server
	<-done
}
