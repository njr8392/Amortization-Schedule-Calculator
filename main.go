package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

var tpl = template.Must(template.ParseFiles("./templates/index.html"))

func index(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

type Loan struct {
	Principal float64
	Interest  float64
	Duration  float64
}
type Amortization struct {
	PrincipalPmt float64
	InterestPmt  float64
	Balance      float64
	Num          int
}
type Schedule []Amortization

func Round(x float64) float64 {
	return math.Ceil(x*100) / 100
}

// loan formula for payment per period
//  A = P (r(1+r)^n)/(1+r)^n -1
//  a = amount per period
// P = Principal amount
// r = interest rate per period
// n = total number of payments or periods
func (l *Loan) monthPmnt() float64 {
	periods := l.Periods()
	return l.Interest * l.Principal / (12 * (1 - (math.Pow(1+l.Interest/12, -periods))))

}
func (l *Loan) Periods() float64 {
	return l.Duration * 12
}
func (l *Loan) Schd() Schedule {
	var intPayment float64
	var prinPmnt float64
	var schd Schedule
	var balance float64 = l.Principal
	monthInt := l.Interest / 12
	monthPmt := l.monthPmnt()

	for i := 1; i <= int(l.Periods()); i++ {

		intPayment = monthInt * balance
		prinPmnt = monthPmt - intPayment
		balance -= prinPmnt

		schd = append(schd, Amortization{PrincipalPmt: Round(prinPmnt), InterestPmt: Round(intPayment), Balance: Round(balance), Num: i})
	}
	return schd
}

func calcHandler(w http.ResponseWriter, r *http.Request) {
	url, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params := url.Query()
	amount, _ := strconv.ParseFloat(params.Get("amount"), 64)
	duration, _ := strconv.ParseFloat(params.Get("duration"), 64)
	interest, _ := strconv.ParseFloat(params.Get("interest"), 64)

	loan := Loan{Principal: amount, Interest: interest, Duration: duration}
	sch := loan.Schd()
	list := template.Must(template.ParseFiles("templates/loan.html"))
	if err := list.Execute(w, sch); err != nil {
		fmt.Printf("%s", err)
	}
}
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/calc", calcHandler)
	http.ListenAndServe("localhost:8000", mux)
}
