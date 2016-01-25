package main

var html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Copy Password</title>
</head>
<body>
	<h1>Copy your password</h1>
	<textarea onclick="this.focus();this.select()" readonly="readonly">{{.}}</textarea>
	
</body>
</html>`
