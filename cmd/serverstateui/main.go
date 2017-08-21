package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/buaazp/fasthttprouter"
	"github.com/henkman/steamquery"
	"github.com/valyala/fasthttp"
	"github.com/zserge/webview"
)

func Players(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.Header.Set("Cache-Control", "No-Cache")
	server := ctx.UserValue("server").(string)
	ps, err := steamquery.QueryPlayersString(server)
	if err != nil {
		ctx.Response.SetBodyString("{'error':" + err.Error() + "}")
		return
	}
	json.NewEncoder(ctx.Response.BodyWriter()).Encode(&ps)
}

func Info(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.Header.Set("Cache-Control", "No-Cache")
	server := ctx.UserValue("server").(string)
	res, err := steamquery.QueryString(server)
	if err != nil {
		ctx.Response.SetBodyString("{'error':" + err.Error() + "}")
		return
	}
	json.NewEncoder(ctx.Response.BodyWriter()).Encode(&res)
}

func Index(server string) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("X-UA-Compatible", "IE=edge")
		ctx.Response.Header.SetContentType("text/html")
		ctx.Response.SetBodyString(`<html>
<head>
	<title>test</title>
	<style>
	html, body { margin:0; padding:0; }
	table.sortable thead {
	    background-color:#eee;
	    color:#666666;
	    font-weight: bold;
	    cursor: default;
	}
	table.sortable td {
		padding: 2px;
	}
	</style>
	<script src="/r/sorttable.js"></script>
	<script src="/r/juration.js"></script>
	<script>
	function updateInfo() {
		var xhr = new XMLHttpRequest();
		xhr.onreadystatechange = function() {
			var DONE = this.DONE || 4;
			if (this.readyState == DONE) {				
				var raw = JSON.parse(this.responseText);
				if(raw.error) {
					alert(raw.error);
					return;
				}
				document.getElementById("name").innerText = raw.Name;
				document.getElementById("map").innerText = raw.Map;
				document.getElementById("playercount").innerText = raw.Players;
			}
		};
		xhr.open("GET", "/info/` + server + `", true);
		xhr.send();
	}
	function updatePlayers() {
		var xhr = new XMLHttpRequest();
		xhr.onreadystatechange = function() {
			var DONE = this.DONE || 4;
			if (this.readyState == DONE) {			
				var raw = JSON.parse(this.responseText);
				if(raw.error) {
					alert(raw.error);
					return;
				}
				var table = document.querySelector("#players");
				var tbody = table.querySelector("tbody");
				tbody.innerHTML = "";
				for(var i=0; i<raw.length; i++) {
					var d = raw[i].Duration;
					var secs = (d/1000000000) + (d%1000000000)/1000000000;
					var tr = document.createElement("tr");

					var td = document.createElement("td");
					td.textContent = raw[i].Name;
					tr.appendChild(td);
					
					td = document.createElement("td");
					td.textContent = raw[i].Score;
					tr.appendChild(td);
					
					td = document.createElement("td");
					td.textContent = juration.stringify(secs);
					td.setAttribute("sorttable_customkey", ''+secs);
					tr.appendChild(td);
					
					var td = document.createElement("td");
					td.textContent = (raw[i].Score/secs).toFixed(6);
					tr.appendChild(td);

					tbody.appendChild(tr);
				}
				var headers = table.querySelectorAll("th");
				for(var i=0; i<headers.length; i++) {
					headers[i].className = headers[i].className
						.replace('sorttable_sorted_reverse','')
						.replace('sorttable_sorted','');
				}
				sorttable.innerSortFunction.apply(headers[3], []);
				sorttable.reverse(tbody);
			}
		};
		xhr.open("GET", "/players/` + server + `", true);
		xhr.send();	
	}
	window.addEventListener("load", function() {		
		var buref = document.getElementById("refresh");
		buref.addEventListener("click", function() {
			updateInfo();
			updatePlayers();
		});
		buref.click();
	});
	</script>
</head>
<body>
	<input type="button" id="refresh" value="refresh"><br/>
	Name: <span id="name"></span><br/>
	Map: <span id="map"></span><br/>
	Players: <span id="playercount"></span><br/>
	<table id="players" class="sortable">
		<thead>
			<tr>
				<th class="sorttable_alpha">Name</th>
				<th class="sorttable_numeric">Score</th>
				<th class="sorttable_numeric">Online</th>
				<th class="sorttable_numeric">Score/Second</th>
			</tr>
		</thead>
    	<tbody></tbody>
	</table>
</body>
</html>`)
	}
}

var (
	_server string
)

func init() {
	flag.StringVar(&_server, "s", "", "server")
	flag.Parse()
}

func main() {
	if _server == "" {
		flag.Usage()
		return
	}
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go func() {
		router := fasthttprouter.New()
		router.GET("/", Index(_server))
		router.GET("/players/:server", Players)
		router.GET("/info/:server", Info)
		router.ServeFiles("/r/*filepath", filepath.Join(filepath.Dir(exe), "r"))
		log.Fatal(fasthttp.Serve(ln, router.Handler))
	}()
	log.Fatal(webview.Open("server state", "http://"+ln.Addr().String(),
		600, 800, false))
}
