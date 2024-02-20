// MIT License
//
// Copyright (c) 2020 The vine Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package web

var (
	layoutTemplate = `
{{define "layout"}}
<html>
	<head>
		<title>Vine Web</title>
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" integrity="sha384-1q8mTJOASx8j1Au+a5WDVnPi2lkFfwwEAa8hDDdjZlpLegxhjVME1fgjWPGmkzs7" crossorigin="anonymous">
		<link href="https://fonts.googleapis.com/css?family=Source+Sans+Pro&display=swap" rel="stylesheet">
		<style>
		  html, body {
		    font-family: 'Source Sans Pro', sans-serif;
		  }
		  html a { color: #333333; }
		  .navbar .navbar-brand { color: #333333; font-weight: bold; font-size: 2.0em; }
		  .navbar-brand img { display: inline; }
		  #navBar, .navbar-toggle { margin-top: 15px; }
		  .icon-bar { background-color: #333333; }
		  .nav>li>a:focus, .nav>li>a:hover { background-color: white; }
                  .navbar-brand.logo {
                        font-size: 3.0em;
                        font-weight: 1000;
                        font-family: medium-content-sans-serif-font,"Lucida Grande","Lucida Sans Unicode","Lucida Sans",Geneva,Arial,sans-serif;
                  }
		 .search {
		    position: relative;
		    max-width: 600px;
		    margin: 0 auto;
		    border-radius: 0;
		    border: 0;
		    box-shadow: none;
		    border-bottom: 1px solid whitesmoke;
		 }
		 .search:focus {
		    border-color: transparent;
		    outline: 0;
		    box-shadow: none;
		    border-bottom: 1px solid whitesmoke;
	 	 }
		 pre {
		    background-color: #fcfcfc;
		    border: 1px solid whitesmoke;
		 }
		 .user {
		    padding: 15px;
		 }
		</style>
		<style>
		{{ template "style" . }}
		</style>
		{{ template "head" . }}
	</head>
	<body>
	  <nav class="navbar">
	    <div class="container">
              <div class="navbar-header">
                <button type="button" class="navbar-toggle" data-toggle="collapse" data-target="#navBar">
                  <span class="icon-bar"></span>
                  <span class="icon-bar"></span>
                  <span class="icon-bar"></span> 
                </button>
                <a class="navbar-brand logo" href="/"><img src="https://vine.mu/logo.png" height=50px width=auto style="margin-bottom: 5px;" /> Vine</a>
              </div>
              <div class="collapse navbar-collapse" id="navBar">
	        <ul class="nav navbar-nav navbar-right" id="dev">
		  {{if gt (len .User) 0 }}<span class="user small">Logged in as: {{.User}}</span>{{end}}
	          <li><a href="/client">Client</a></li>
	          <li><a href="/services">Services</a></li>
	          {{if .StatsURL}}<li><a href="{{.StatsURL}}" class="navbar-link">Stats</a></li>{{end}}
	          {{if .LoginURL}}<li><a href="{{.LoginURL}}" class="navbar-link">{{.LoginTitle}}</a></li>{{end}}
	        </ul>
              </div>
	    </div>
	  </nav>
          <div class="container">
            <div class="row">
	      <div class="col-sm-12">
                {{ template "heading" . }}
                {{ template "content" . }}
              </div>
            </div>
          </div>
	  <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.1.4/jquery.min.js"></script>
	  <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js" integrity="sha384-0mSbJDEHialfmuBBQP6A4Qrprq5OVfW37PRR3j5ELqxss1yVqOtnepnHVP9aJ7xS" crossorigin="anonymous"></script>
	  {{template "script" . }}
	  <script type="text/javascript">
		function toggle(e) {
		      var ev = window.event? event : e
		      if (ev.keyCode == 80 && ev.ctrlKey && ev.shiftKey) {
			var el = document.getElementById("dev");
			if (el.style.display == "none") {
			  el.style.display = "block";
			} else {
			  el.style.display = "none";
			}
		    }
		}

		document.onkeydown = toggle;
	  </script>
	</body>
</html>
{{end}}
{{ define "style" }}
.service { border-radius: 100px; }
{{end}}
{{ define "head" }}{{end}}
{{ define "script" }}{{end}}
{{ define "title" }}{{end}}
{{ define "heading" }}<h3>&nbsp;</h3>{{end}}
`

	indexTemplate = `
{{define "heading"}}<h4><input class="form-control input-lg search" type=text placeholder="Search" autofocus></h4>{{end}}
{{define "style" }}
.search {
  border-radius: 0;
  border: 0;
  box-shadow: none;
  border-bottom: 1px solid whitesmoke;
}
.search:focus {
  border-color: transparent;
  outline: 0;
  box-shadow: none;
  border-bottom: 1px solid whitesmoke;
}
.service {
	margin: 5px 3px 5px 3px;
	padding: 20px;
	text-align: center;
	display: block;
}
.search { border-radius: 100px; }
.apps {
  max-width: 600px;
  text-align: center;
  margin: 0 auto;
}
.icon {
  width: 60px;
  height: 60px;
  display: block;
  border-radius: 50px; 
  border: 2px solid whitesmoke;
  background-color: #fcfcfc;
  color: #AFACBE;
  font-size: 40px;
  font-weight: bold;
}
.icon:hover {
  color: #23527c;
  border: 2px solid #23527c;
}
.apps .service:hover {
  text-decoration: none;
  font-weight: bold;
}
@media only screen and (max-width: 500px) {
  .service {
    padding: 5px;
  }
}
{{end}}
{{define "title"}}Web{{end}}
{{define "content"}}
	{{if .Results.HasWebServices}}
		<div class="apps">
			{{range .Results.WebServices}}
			<div style="display: inline-block; max-width: 150px; vertical-align: top;">
			<a href="{{.Link}}" data-filter={{.Name}} class="service">
			  <div style="padding: 5px; max-width: 80px; display: block; margin: 0 auto;">
				{{if .Icon }}<img src="{{.Icon}}" style="width: 70px; height: auto;"/>{{else}}
				<div class="icon">{{First .Name}}</div>
				{{end}}
			  </div>
			  <div>{{Title .Name}}</div>
			</a>
			</div>
			{{end}}
		</div>
	{{end}}
{{end}}
{{define "script"}}
<script type="text/javascript">
jQuery(function($, undefined) {
	var refs = $('a[data-filter]');
	$('.search').on('keyup', function() {
		var val = $.trim(this.value);
		refs.hide();
		refs.filter(function() {
			return $(this).data('filter').search(val) >= 0
		}).show();
	});

	$('.search').on('keypress', function(e) {
		if (e.which != 13) {
			return;
		}
		$('.service').each(function() {
			if ($(this).css('display') == "none") {
				return;
			}
			window.location.href = $(this).attr('href');
		})
	});
});

</script>
{{end}}
`
	callTemplate = `
{{define "title"}}Call{{end}}
{{define "style"}}
	pre {
		word-wrap: break-word;
		border: 0;
	}
	.form-control {
		border: 1px solid whitesmoke;
	}
{{end}}
{{define "content"}}
<div class="row">
  <div class="panel">
    <div class="panel-body">
	<div class="col-sm-5">
		<form id="call-form" onsubmit="return call();">
			<div class="form-group">
				<label for="service">Service</label>
				<ul class="list-group">
					<select class="form-control" type=text name=service id=service> 
					<option disabled selected> -- select a service -- </option>
					{{range $key, $value := .Results}}
					<option class = "list-group-item" value="{{$key}}">{{$key}}</option>
					{{end}}
					</select>
				</ul>
			</div>
			<div class="form-group">
				<label for="endpoint">Endpoint</label>
				<ul class="list-group">
					<select class="form-control" type=text name=endpoint id=endpoint>
					<option disabled selected> -- select an endpoint -- </option>
					</select>
				</ul>
			</div>
			<div class="form-group other" style="display: none;">
				<label for="otherendpoint">Other Endpoint</label>
				<ul class="list-group">
					<input class="form-control" type=text name=otherendpoint id=otherendpoint placeholder="Endpoint"/>
				</ul>
			</div>
			<div class="form-group">
			</div>
			<div class="form-group">
				<label for="metadata">Metadata</label>
				<ul class="list-group">
					<input class="form-control" type=text name=metadata id=metadata placeholder="Metadata" value="{}"/>
				</ul>
				<label for="request">Request</label>
				<textarea class="form-control" name=request id=request rows=8>{}</textarea>
			</div>
			<div class="form-group">
				<button class="btn btn-default" style="border-color: whitesmoke;">Call</button>
			</div>
		</form>
	</div>
	<div class="col-sm-7">
		<p><b>Response</b><span class="pull-right"><a href="#" onclick="copyResponse()">Copy</a></p>
		<pre id="response" style="min-height: 405px; max-height: 405px; overflow: scroll;">{}</pre>
	</div>
    </div>
  </div>
</div>
{{end}}
{{define "script"}}
	<script>
		function copyResponse() {
			var copyText = document.getElementById("response");
			const textArea = document.createElement('textarea');
			textArea.textContent = copyText.innerText;
			textArea.style = "position: absolute; left: -1000px; top: -1000px";	
			document.body.append(textArea);
			textArea.select();
			textArea.setSelectionRange(0, 99999);
			document.execCommand("copy");
			document.body.removeChild(textArea);
			return false;
		}
	</script>
	<script>
		$(document).ready(function(){
			//Function executes on change of first select option field 
			$("#service").change(function(){
				var select = $("#service option:selected").val();
				$("#endpoint").empty();
				$("#endpoint").append("<option disabled selected> -- select an endpoint -- </option>");
				var s_map = {};
				{{ range $service, $endpoints := .Results }}
				var m_list = [];
				{{range $index, $element := $endpoints}}
				m_list[{{$index}}] = {{$element.Name}}
				{{end}}
				s_map[{{$service}}] = m_list
				{{ end }}
				if (select in s_map) {
					var serviceEndpoints = s_map[select]
					var len = serviceEndpoints.length;
					for(var i = 0; i < len; i++) {
						$("#endpoint").append("<option value=\""+serviceEndpoints[i]+"\">"+serviceEndpoints[i]+"</option>");	
					}
				}
				$("#endpoint").append("<option value=\"other\"> - Other</option>");
			});

			//Function executes on change of second select option field 
			$("#endpoint").change(function(){
				var select = $("#endpoint option:selected").val();
				if (select == "other") {
					$(".other").css('display', 'block');
					$("#otherendpoint").attr("disabled", false);
				} else {
					$(".other").css('display', 'none');
					$("#otherendpoint").attr("disabled", true);
					$('#otherendpoint').val('');
				}

			});
		});
	</script>
	<script>
		function call() {
			var req = new XMLHttpRequest()
			req.onreadystatechange = function() {
				if(req.readyState != 4) {
					return
				}
				if (req.readyState == 4 && req.status == 200) {
					document.getElementById("response").innerText = JSON.stringify(JSON.parse(req.responseText), null, 2);
				} else if (req.responseText.slice(0, 1) == "{") {
					document.getElementById("response").innerText = JSON.stringify(JSON.parse(req.responseText), null, 2);
				} else if (req.responseText.length > 0) {
					document.getElementById("response").innerText = req.responseText;
				} else {
					document.getElementById("response").innerText = "Request error " + req.status;
				}
				console.log(req.responseText);
			}
			var endpoint = document.forms[0].elements["endpoint"].value
			if (!($('#otherendpoint').prop('disabled'))) {
				endpoint = document.forms[0].elements["otherendpoint"].value
			}

			var reqBody;
			var headers;

			try {
				var md = document.forms[0].elements["metadata"].value;
				var rq = document.forms[0].elements["request"].value
				if (md.length > 0) {
					headers = JSON.parse(md);
				}
				if (rq.length > 0) {
					reqBody = JSON.parse(rq);
				};
			} catch(e) {
				document.getElementById("response").innerText = "Invalid request: " + e.message;
				return false;
			}

			var request = {
				"service": document.forms[0].elements["service"].value,
				"endpoint": endpoint,
				"request": reqBody
			}
			req.open("POST", "/rpc", true);
			req.setRequestHeader("Content-type","application/json");

			if (headers != undefined) {
				for (let [key, value] of Object.entries(headers)) {
					req.setRequestHeader(key, value);
				}
			}

			req.send(JSON.stringify(request));

			return false;
		};	
	</script>
{{end}}
`
	registryTemplate = `
{{define "heading"}}<h4><input class="form-control input-lg search" type=text placeholder="Search" autofocus></h4>{{end}}
{{define "title"}}Services{{end}}
{{define "content"}}
	<p style="margin: 0;">&nbsp;</p>
        <div style="max-width: 600px; margin: 0 auto; height: calc(100vh - 200px); overflow: scroll;">
	{{range .Results}}
	<div style="margin: 5px 5px 5px 15px;">
	    <a href="/service/{{.Name}}" data-filter={{.Name}} class="service">{{.Name}}</a>
	</div>
	{{end}}
        </div>
{{end}}
{{define "script"}}
<script type="text/javascript">
jQuery(function($, undefined) {
	var refs = $('a[data-filter]');
	$('.search').on('keyup', function() {
		var val = $.trim(this.value);
		refs.hide();
		refs.filter(function() {
			return $(this).data('filter').search(val) >= 0
		}).show();
	});
});
</script>
{{end}}
`

	serviceTemplate = `
{{define "title"}}Service{{end}}
{{define "heading"}}<h3>{{with $svc := index .Results 0}}{{$svc.Name}}{{end}}</h3>{{end}}
{{define "style"}}
.table>tbody>tr>th, .table>tbody>tr>td {
    border-top: none;
}
.endpoint {
  cursor: pointer;
}
.bold {
  font-weight: bold;
}
pre {padding: 20px;}
{{end}}
{{define "script"}}
<script type="text/javascript">
  $('.endpoint').on('click', function() {
	var val = $(this).parent().find("table");
	var state = $(this).find(".state");
	if (val.css('display') == 'none') {
	  state.text("[-]");
	  val.css('display', 'table');
	} else {
	  val.css('display', 'none');
	  state.text("[+]");
	}
  });
</script>
{{end}}
{{define "content"}}
	<hr>
	<h4 class="bold">Nodes</h4>
	{{range .Results}}
	<h5>Version: {{.Version}}</h5>
	<table class="table">
		<thead>
			<th>Id</th>
			<th>Address</th>
			<th>Metadata</th>
		<thead>
		<tbody>
			{{range .Nodes}}
			<tr>
				<td>{{.Id}}</td>
				<td>{{.Address}}</td>
				<td>{{ range $key, $value := .Metadata }}{{$key}}={{$value}} {{end}}</td>
			</tr>
			{{end}}
		</tbody>
	</table>
	{{end}}
	{{with $svc := index .Results 0}}
	{{if $svc.Endpoints}}
	<h4 class="bold">Endpoints</h4>
	<hr/>
	{{end}}
	{{range $svc.Endpoints}}
	<div>
		<h4 class="endpoint"><span class="state">[+]</span> {{.Name}}</h4>
		<table class="table" style="display: none;">
			<tbody>
				{{if .Metadata}}
				<tr>
					<th class="col-sm-2" scope="row">Metadata</th>
					<td>{{ range $key, $value := .Metadata }}{{$key}}={{$value}} {{end}}</td>
				</tr>
				{{end}}
				<tr>
					<th class="col-sm-2" scope="row">Request</th>
					<td><pre>{{format .Request}}</pre></td>
				</tr>
				<tr>
					<th class="col-sm-2" scope="row">Response</th>
					<td><pre>{{format .Response}}</pre></td>
				</tr>
			</tbody>
		</table>
	</div>
	{{end}}
	{{end}}
{{end}}

`
)
