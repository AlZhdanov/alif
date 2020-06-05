package main

import (
	"fmt"
	"database/sql"
    "net/http"
    "encoding/json"
    "log"
    "net"
    "os"
    "net/smtp"
    "net/mail"
    "crypto/tls"
    _ "github.com/mattn/go-sqlite3"
)

var database  *sql.DB

type buyer struct{
	Id int        `json:"id"`
	Phone string  `json:"phone"`
	Email string  `json:"email"` 
}

func sendMessage(emailTo, body string) {
    from := mail.Address{"", "alif.tech.zh@mail.ru"}
    to   := mail.Address{"", emailTo}
    subj := "info"
    pass, exists := os.LookupEnv("PASS")
    
    if !exists {
        log.Panic("Password not set")
        return
    }
    
    headers := make(map[string]string)
    headers["From"] = from.String()
    headers["To"] = to.String()
    headers["Subject"] = subj

    message := ""
    for k,v := range headers {
        message += fmt.Sprintf("%s: %s\r\n", k, v)
    }
    message += "\r\n" + body

    servername := "smtp.mail.ru:465"
    host, _, _ := net.SplitHostPort(servername)
    

	
    auth := smtp.PlainAuth("", "alif.tech.zh@mail.ru", pass, host)

	tlsconfig := &tls.Config {
        InsecureSkipVerify: true,
        ServerName: host,
    }

    conn, err := tls.Dial("tcp", servername, tlsconfig)
    if err != nil {
        log.Panic(err)
    }

    c, err := smtp.NewClient(conn, host)
    if err != nil {
        log.Panic(err)
    }

    if err = c.Auth(auth); err != nil {
        log.Panic(err)
    }

    // To && From
    if err = c.Mail(from.Address); err != nil {
        log.Panic(err)
    }

    if err = c.Rcpt(to.Address); err != nil {
        log.Panic(err)
    }

    // Data
    w, err := c.Data()
    if err != nil {
        log.Panic(err)
    }

    _, err = w.Write([]byte(message))
    if err != nil {
        log.Panic(err)
    }

    err = w.Close()
    if err != nil {
        log.Panic(err)
    }

    c.Quit()

}

func addBuyer(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST" {
        w.Write([]byte("ERROR, expected method POST"))
    }
    r.ParseForm()
    phone := r.FormValue("phone") 
    email := r.FormValue("email") 

    _, err := database.Exec("insert into buyer (phone, email) values ( $1, $2)", 
    phone, email)
    if err != nil{
        panic(err)
    }

    w.Write([]byte("Ok"))
}

func getBuyer(w http.ResponseWriter, r *http.Request){
	rows, err := database.Query("select id, phone, email from buyer")
    if err != nil {
        panic(err)
    }
	buyers := []buyer{}
     
    for rows.Next(){
        p := buyer{}
        err := rows.Scan(&p.Id, &p.Phone, &p.Email)
        if err != nil{
            fmt.Println(err)
            continue
        }
        buyers = append(buyers, p)
    }
    
    w.Header().Set("Content-Type", "application/json") 
    json.NewEncoder(w).Encode(buyers) 
}

func sendInfo(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        w.Write([]byte("ERROR, expected method POST"))
    }
    r.ParseForm()
    idBuyer := r.FormValue("id")
    methoSending  := r.FormValue("methoSending")
    body := r.FormValue("body")
    

    rows, err := database.Query("select id, phone, email from buyer where id=$1", idBuyer)
    if err != nil {
        panic(err)
    }
	buyers := []buyer{}
     
    for rows.Next(){
        p := buyer{}
        err := rows.Scan(&p.Id, &p.Phone, &p.Email)
        if err != nil{
            fmt.Println(err)
            continue
        }
        buyers = append(buyers, p)
    }

    if methoSending == "email" {
        sendMessage(buyers[0].Email, body)
    }

    if methoSending == "sms" {
        w.Write([]byte("gate not found"))
    }
}

func main()  {

	db, err := sql.Open("sqlite3", "test.db")
    if err != nil {
        panic(err)
	}
        
    database = db
	defer db.Close()

    http.HandleFunc("/add-buyer", addBuyer)
    http.HandleFunc("/get-buyer", getBuyer)
    http.HandleFunc("/send-info", sendInfo)
    fmt.Println("server is running on the port 8080")
	http.ListenAndServe("localhost:8080", nil)
}