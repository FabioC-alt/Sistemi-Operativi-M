package main

import (
	"fmt"
	"math/rand"
	"time"
)

const Nord = 1
const Sud = 0
const Auto = 1
const Pedone = 0
const MAX = 30
const MAXBUFF = 200
const MAXPROC = 100

var done = make(chan bool)
var termina = make(chan bool)
var canaleAutoInN = make(chan int)
var canaleAutoInS = make(chan int)
var canalePedInN = make(chan int)
var canalePedInS = make(chan int)
var canaleAutoOutN = make(chan int)
var canaleAutoOutS = make(chan int)
var canalePedOutN = make(chan int)
var canalePedOutS = make(chan int)
var ack [MAXPROC]chan int

func veicolo(myid int, dir int, tipo int) {

	var tt int
	tt = rand.Intn(5) + 1
	time.Sleep(time.Duration(tt) * time.Second)
	fmt.Println("Inizio simulazione")
	if dir == Nord {
		if tipo == Auto {
			canaleAutoInN <- myid
			<-ack[myid]
			fmt.Println("Auto con id:%d entra da Nord", myid)
			tt = rand.Intn(5)
			time.Sleep(time.Duration(tt) * time.Second)
			fmt.Println("Auto con id:%d esce da Sud(N)", myid)
			canaleAutoOutN <- myid
		} else {
			canalePedInN <- myid
			<-ack[myid]
			fmt.Println("Pedone con id:%d entra da Nord", myid)
			tt = rand.Intn(5)
			time.Sleep(time.Duration(tt) * time.Second)
			fmt.Println("Pedone con id:%d esce da Sud(N)", myid)
			canalePedOutN <- myid
		}
	} else {
		if tipo == Auto {
			canaleAutoInS <- myid
			<-ack[myid]
			fmt.Println("Auto con id:%d entra da Sud", myid)
			tt = rand.Intn(5)
			time.Sleep(time.Duration(tt) * time.Second)
			fmt.Println("Auto con id:%d esce da Nord(S)", myid)
			canaleAutoOutS <- myid
		} else {
			canalePedInS <- myid
			<-ack[myid]
			fmt.Println("Pedone con id:%d entra da Sud", myid)
			tt = rand.Intn(5)
			time.Sleep(time.Duration(tt) * time.Second)
			fmt.Println("Pedone con id:%d esce da Nord(N)", myid)
			canalePedOutS <- myid
		}

	}

	done <- true
}

func server() {
	fmt.Println("Server inizializzato")
	var countAN int = 0
	var countAS int = 0
	var countPN int = 0
	var countPS int = 0

	for {

		select {
		case x := <-when((countAS+countPN+countPS < MAX && countAN == 0), canalePedInS):
			fmt.Println("Server: Servito pedone con id: %d entra da Sud", x)
			countPS++
			ack[x] <- 1
		case x := <-when((countPN+countPS+countAN < MAX && countAS == 0) && (len(canalePedInS) == 0), canalePedInN):
			fmt.Println("Server: Servito pedone con id: %d entra da Nord", x)
			countPN++
			ack[x] <- 1
		case x := <-when((countAS+countPS+10 < MAX && countAN == 0) && (len(canalePedInN) == 0 && len(canalePedInS) == 0), canaleAutoInS):
			fmt.Println("Server: Servita Auto con id: %d entra da Sud", x)
			countAS = countAS + 10
			ack[x] <- 1
		case x := <-when((countAN+countPN+10 < MAX && countAS == 0) && (len(canalePedInN) == 0 && len(canalePedInS) == 0 && len(canaleAutoInS) == 0), canaleAutoInN):
			fmt.Println("Server: Servita Auto con id: %d entra da Nord", x)
			countAN = countAN + 10
			ack[x] <- 1
		case x := <-canalePedOutS:
			fmt.Println("Server: Servito pedone con id: %d uscito da Sud", x)
			countPS--
		case x := <-canalePedOutN:
			fmt.Println("Server: Servito pedone con id: %d uscito da Nord", x)
			countPN--
		case x := <-canaleAutoOutS:
			fmt.Println("Server: Servita Auto con id: %d ucito da Sud", x)
			countAS = countAS - 10
		case x := <-canaleAutoOutN:
			fmt.Println("Server: Servita Auto con id:%d uscito da Nord", x)
			countAN = countAN - 10
		case <-termina:
			done <- true
			return
		}
	}
}

func main() {

	fmt.Println("Inizio simulazione")
	var AN int = 10
	var AS int = 10
	var PN int = 20
	var PS int = 20

	for i := 0; i < AN+AS+PN+PS; i++ {

		ack[i] = make(chan int, MAXBUFF)

	}

	rand.Seed(time.Now().Unix())

	go server()

	for i := 0; i < AN; i++ {

		go veicolo(i, Auto, Nord)
	}

	for i := 0; i < AS; i++ {

		go veicolo(i, Auto, Sud)
	}
	for i := 0; i < PN; i++ {

		go veicolo(i, Pedone, Nord)
	}
	for i := 0; i < PS; i++ {

		go veicolo(i, Pedone, Sud)
	}

	for i := 0; i < AN+AS+PS+PN; i++ {
		<-done
	}
	termina <- true
	<-done
	fmt.Println("Simulazione Terminata")
}

func when(b bool, c chan int) chan int {

	if !b {
		return nil
	}
	return c

}
