package splopi


import (
        "html/template"
        "net/http"
)


func init() {
        http.HandleFunc("/roc/*", register)
        http.HandleFunc("/roc/register", register)
}

// [START func_root]
func register(w http.ResponseWriter, r *http.Request) {

        if err := registerTemplate.Execute(w , nil); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}

var registerTemplate = template.Must(template.New("register").Parse(`
<html>
  <head>
    <title>Time to register</title>
  </head>
  <body>
    <form action="/sign" method="post">
      <div><input type="text" name="name" placeholder="Name"></div>
      <div><input type="password" placeholder="Password"></div>
      <div><input type="password" placeholder="Re enter password"></div>
      <div><input type="text" placeholder="Challenge"></div>
      <div><input type="submit" value="Register"></div>
    </form>
  </body>
</html>
`))
