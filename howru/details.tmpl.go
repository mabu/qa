package main

import "html/template"

var tmplDetails = template.Must(template.New("details").Parse(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en">
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<title>howru {{.HWAddrs}}</title>
	</head>
	<body>
		<table>
			<caption>{{.HWAddrs}}</caption>
			<tr>
				<th>Tinklo adresai</th>
				<th>IP adresas</th>
				<th>Veikimo laikas</th>
				<th>Apkrova</th>
				<th>Laikas</th>
				<th>Serverio laikas</th>
				<th>Klaidos</th>
			</tr>
			{{range .Data}}
			<tr>
				<td>{{.Addrs}}</td>
				<td>{{.IP}}</td>
				<td>{{.Uptime}}</td>
				<td>{{.Loads}}</td>
				<td>{{.Time}}</td>
				<td>{{.SrvTime}}</td>
				<td>{{.Errors}}</td>
			</tr>
			{{end}}
		</table>
	</body>
</html>`))
