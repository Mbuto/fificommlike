package hello

import (
    "fmt"
    "time"
    "net/http"
    "appengine"
    "appengine/urlfetch"
    "appengine/datastore"
    "encoding/json"
    "io/ioutil"
)

func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/help", help)
    http.HandleFunc("/sign", sign)
    http.HandleFunc("/robots.txt", robots)
}

func root(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, mioForm)
}

const mioForm = `
<html>
  <body>
    <h1>Cruscotto FF</h1>
    <h2>Commenti e Like sugli ultimi 10 post</h2>
    <form action="/sign" method="post">
	<table>
      <tr><td>Username FF:</td><td><input type="text" name="user"></td></tr>
      <tr><td>Remote Key FF:</td><td><input type="text" name="pw"></td></tr>
      <tr><td colspan=2><a href="http://friendfeed.com/remotekey" target="_blank">Se non ce l&apos;hai, ottieni la tua Remote Key qui</a></td></tr>
      <tr><td>Target User:</td><td><input type="text" name="tgt"></td></tr>
	</table>
      <p><div><input type="submit" value="Invio">
      <p><b>Hai letto le istruzioni?</b> <a href="/help">Help/Info</a></div>
<p><img src="https://developers.google.com/appengine/images/appengine-silver-120x30.gif" 
alt="Powered by Google App Engine" />
    </form>
<hr><div><b>Disclaimer:</b> Questa applicazione utilizza un servizio <i>gratuito</i> soggetto a <b>limitazioni</b>.<br>Il programma viene reso disponibile <b>senza alcuna garanzia</b> di corretto funzionamento:<br>utilizzatelo liberamente a vostro rischio.</div>
<p><hr><div>Grazie per aver usato questo programma.<br>
Eventuali critiche o suggerimenti possono essere diretti a:
<ul>
<li>Gruppo <a target="_blank" href="http://friendfeed.com/ff-buoni-e-cattivi">"FF Buoni &amp; Cattivi"</a>
</ul></div>
  </body>
</html>
`

const helpForm = `
<html>
  <body>
    <h1>Cruscotto FF</h1>
<div>Questa App visualizza un cruscotto con informazioni relative a:
<ul><li>numero di commenti<li> numero di like</ul>sugli <b>ultimi 10 post del proprio Feed<br>o del Feed di qualcuno a cui siamo iscritti</b> (Target User) su FriendFeed.</div>
<hr>
Per ulteriori info:
<ul>
<li>Gruppo <a target="_blank" href="http://friendfeed.com/ff-buoni-e-cattivi">"FF Buoni &amp; Cattivi"</a>
</ul>
<div><a href="/">Ritorna alla Home Page</a></div>
<p><img src="https://developers.google.com/appengine/images/appengine-silver-120x30.gif" 
alt="Powered by Google App Engine" />
    </form>
  </body>
</html>
`

func help(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, helpForm)
}


type Utente struct {
Id string
Lastaccess time.Time
Commts int
Liks int
}

func sign(w http.ResponseWriter, r *http.Request) {

    s := r.FormValue("user")
    t := r.FormValue("pw")
    g := r.FormValue("tgt")

if g == "" {
    g = s
}

kk := controlla(s, t, g, w, r)
if kk > 1 {
	fmt.Fprintf(w, "<html><body><br><br><br>")
	fmt.Fprintf(w, "<b> ATTENZIONE! errore FriendFeed (%d) </b>", kk)
	fmt.Fprintf(w, "<br>Username: %s",s)
	fmt.Fprintf(w, "<br>Target User: %s",g)
	fmt.Fprintf(w, "<br><br><a href=/>Home</a></body></html>")
	return
}
if kk > 0 {
  fmt.Fprintf(w, "<b>Username o RemoteKey errato</b> <a href=/>Riprova</a>")
	return
}

return
}

func controlla(u string, pw string, tgt string, ww http.ResponseWriter, rr *http.Request) int {
c := appengine.NewContext(rr)
    client := urlfetch.Client(c)
z := "http://friendfeed-api.com/v2/feed/" + tgt + "?num=10"
 req2, err2 := http.NewRequest("GET", z, nil)
    if err2 != nil {
        http.Error(ww, err2.Error(), http.StatusInternalServerError)
	return 1
    }
req2.SetBasicAuth(u,pw)
resp2, err2 := client.Do(req2)
    if err2 != nil {
//        http.Error(ww, err2.Error(), http.StatusInternalServerError)
    	return resp2.StatusCode
    }
    if resp2.Status == "401 Unauthorized" {
    return 1
    }
    if resp2.Status != "200 OK" {
//    fmt.Fprintf(ww, "client.Do ERR:%v ", resp2)
    return resp2.StatusCode
    }
body, err4 := ioutil.ReadAll(resp2.Body)
    if err4 != nil {
	fmt.Fprintf(ww, "OHHO Err.ReadAll %v\n", err4)
	return 1
    }

resp2.Body.Close()

var animals Ani

	err3 := json.Unmarshal(body, &animals)
	if err3 != nil {
		fmt.Fprintf(ww, "Unmarshal error: %v\n", err3)
	return 1
	}
nn := 0
mm := 0
for kk := 0; kk < len(animals.Entries); kk++ {
  for jj := 0; jj < len(animals.Entries[kk].Comments); jj++ {
    if animals.Entries[kk].Comments[jj].From.Id != tgt {
	nn++
    }
  }
  for jj := 0; jj < len(animals.Entries[kk].Likes); jj++ {
    if animals.Entries[kk].Likes[jj].From.Id != tgt {
	mm++
    }
  }
}

fmt.Fprintf(ww, gaugeForm1, nn, mm)

uu := Utente{tgt,time.Now(),nn,mm}
_, err5 := datastore.Put(c, datastore.NewIncompleteKey(c, "Utente", nil), &uu)
    if err5 != nil {
	fmt.Fprintf(ww, "<p>err5=%v\n",err5)
fmt.Fprintf(ww, "<hr><div><b>Tip &amp; Tricks:</b> <i>Errori \"Over Quota\"? Riprovate la mattina dopo le 9:00</i> </div>")
	return 1
    }

var zz []*Utente

qq := datastore.NewQuery("Utente").Filter("Id =", tgt).Order("-Lastaccess").Limit(10)
_, err6 := qq.GetAll(c, &zz)
	if err6 != nil {
		fmt.Fprintf(ww, "err6=%v\n",err6)
		return 1
	}
if len(zz) < 2 {
	fmt.Fprintf(ww, gaugeForm2, "Never", 0, 0)
} else {
  for yy := (len(zz)-1); yy >= 0; yy-- {
	fmt.Fprintf(ww, gaugeForm2, zz[yy].Lastaccess.Format(time.ANSIC), zz[yy].Commts, zz[yy].Liks)
  }
}
fmt.Fprintf(ww, gaugeForm3, tgt, time.Now().Format(time.ANSIC), u, pw, tgt)
return 0
}

const gaugeForm1 = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8"/>
    <title>
      Cruscotto FF
    </title>
    <script type="text/javascript" src="http://www.google.com/jsapi"></script>
    <script type="text/javascript">
      google.load('visualization', '1', {packages: ['corechart','gauge']});
    </script>
    <script type="text/javascript">
      function drawVisualization() {
        // Create and populate the data table.
var data2 = google.visualization.arrayToDataTable([
          ['Label', 'Value'],
['Commenti', %d], ['Like', %d]
        ]);

var data = google.visualization.arrayToDataTable([
          ['Data-ora', 'Commenti', 'Like'],
`

const gaugeForm2 = `
	['%s',   %d,  %d, ],
`

const gaugeForm3 = `
        ]);

        var options = {
          width: 400, height: 120,
          redFrom: 150, redTo: 200,
          yellowFrom: 100, yellowTo: 150,
          minorTicks: 5,
	  max: 200
        };
      
	var gaug = new google.visualization.Gauge(document.getElementById('gaug_div'));
	    gaug.draw(data2, options);

        new google.visualization.LineChart(document.getElementById('visualization')).
            draw(data, {curveType: "function",
                        width: 800, height: 400,
                        vAxis: {maxValue: 10, gridlines: {count: 10}}}
                );
      }
      
    google.setOnLoadCallback(drawVisualization);
    </script>
  </head>
  <body style="font-family: Arial;border: 0 none;">
<div align=center>
<h1>Commenti e Like per %s</h1>
    <div id="gaug_div" style="width: 800px; height: 150px;"></div>
<h2>Agg.to: %s UTC</h2>
    <div id="visualization" style="width: 800px; height: 400px;"></div>
</div>
<div><a href=/sign?user=%s&pw=%s&tgt=%s>Refresh</a>
  <br><a href=/>Home</a></div>
  </body>
</html>
`


func checkuser(u string, pw string, ww http.ResponseWriter, rr *http.Request) int {

c := appengine.NewContext(rr)
    client := urlfetch.Client(c)
 req, err := http.NewRequest("GET", "http://friendfeed-api.com/v2/validate", nil)
    if err != nil {
	fmt.Fprintf(ww, "http.NewRequest ERROR:")
        http.Error(ww, err.Error(), http.StatusInternalServerError)
    }
req.SetBasicAuth(u,pw)
resp, err := client.Do(req)
    if err != nil {
	fmt.Fprintf(ww, "client.Do1 ERROR:")
        http.Error(ww, err.Error(), http.StatusInternalServerError)
	return 1
    }
    if resp.Status == "401 Unauthorized" {
	return 1
}
    if resp.Status != "200 OK" {
       fmt.Fprintf(ww, "client.Do2 ERROR:%v\n", resp.Status)
       if resp.Status == "401 Unauthorized" {
       fmt.Fprintf(ww, "<b>Username o RemoteKey errato</b> <a href=/>Riprova</a>")
       return 1
       }
        fmt.Fprint(ww, "ERRORE GENERICO: %v <a href=/>Riprova</a>", resp)
    return 1
    }
return 0
}

func robots(w http.ResponseWriter, r *http.Request) {
fmt.Fprintf(w, "User-agent: *\nDisallow: /\n")
}

type Fromt struct {
Id string
}

type Commt struct {
From Fromt
}

type Entt struct {
Comments []Commt
Likes []Commt
}

type Ani struct {
Entries []Entt
}

