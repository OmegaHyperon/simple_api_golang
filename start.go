package main

import (
	"conf_util"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"ora_conn"
	"os"
	"strconv"
	"sync"

	_ "github.com/godror/godror"
)

type application struct {
	counter  int
	mutex    *sync.Mutex
	conf     *conf_util.ConfUtil
	errorLog *log.Logger
	infoLog  *log.Logger
}

func (a *application) headers(req *http.Request) string {
	var res string = ""

	for name, headers := range req.Header {
		for _, h := range headers {
			res += fmt.Sprintf("%v: %v\n", name, h)
		}
	}

	return res
}

func (a *application) headersString(w http.ResponseWriter, r *http.Request) {
	a.infoLog.Printf(">>> headersString")
	w.Header().Set("MYHEADER", "AAABBBCCCDDDEEEFFFGGG")
	fmt.Fprintf(w, "List of headers: \n\n")
	fmt.Fprintln(w, a.headers(r))
}

func (a *application) echoString(w http.ResponseWriter, r *http.Request) {
	a.infoLog.Printf(">>> echoString")
	w.Header().Set("MYHEADER", "AAABBBCCCDDDEEEFFFGGG")
	fmt.Fprintf(w, "hello\n\n")
}

func (a *application) incrementCounter(w http.ResponseWriter, r *http.Request) {
	a.infoLog.Printf(">>> incrementCounter")
	w.Header().Set("MYHEADER", "AAABBBCCCDDDEEEFFFGGG")
	a.mutex.Lock()
	a.counter++
	fmt.Fprintf(w, strconv.Itoa(a.counter))
	a.mutex.Unlock()
}

func (a *application) randDigit(w http.ResponseWriter, r *http.Request) {
	a.infoLog.Printf(">>> randDigit")
	w.Header().Set("MYHEADER", "AAABBBCCCDDDEEEFFFGGG")
	fmt.Fprintf(w, strconv.Itoa(rand.Intn(100)))
}

func (a *application) test_oracle() {
	ora_conn.ConnectToOracle(a.conf.DB_username, a.conf.DB_password, a.conf.DB_conn)
	fmt.Println(">>> End of ConnectToOracle")
}

func main() {
	appl := application{}

	f_log, err := os.OpenFile("./logs/simple_api_golang.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f_log.Close()

	mwh := io.MultiWriter(os.Stdout, f_log)

	// Используйте log.New() для создания логгера для записи информационных сообщений. Для этого нужно
	// три параметра: место назначения для записи логов (os.Stdout), строка
	// с префиксом сообщения (INFO или ERROR) и флаги, указывающие, какая
	// дополнительная информация будет добавлена. Обратите внимание, что флаги
	// соединяются с помощью оператора OR |.
	appl.infoLog = log.New(mwh, "INFO\t", log.Ldate|log.Ltime)

	// Создаем логгер для записи сообщений об ошибках таким же образом, но используем stderr как
	// место для записи и используем флаг log.Lshortfile для включения в лог
	// названия файла и номера строки где обнаружилась ошибка.
	appl.errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	appl.mutex = &sync.Mutex{}
	appl.counter = 0

	appl.infoLog.Printf("Start of the server...")
	// appl.errorLog.Printf("No errors at start!")

	appl.conf = &conf_util.ConfUtil{}
	appl.conf.LoadIniFile()

	appl.infoLog.Printf(">>> Test of Oracle...")
	appl.test_oracle()

	// Используем методы из структуры в качестве обработчиков маршрутов.
	mux := http.NewServeMux()
	mux.HandleFunc("/headers", appl.headersString)
	mux.HandleFunc("/inc", appl.incrementCounter)
	mux.HandleFunc("/rand", appl.randDigit)
	mux.HandleFunc("/", appl.echoString)

	var serv_url string = fmt.Sprintf(":%d", appl.conf.Port)
	addr := flag.String("addr", serv_url, "Сетевой адрес веб-сервера")
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: appl.errorLog,
		Handler:  mux,
	}
	fmt.Println(">>> ", serv_url, appl.conf.Port)
	errserv := srv.ListenAndServe()
	appl.errorLog.Fatal(errserv)
}
