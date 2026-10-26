package main // Declaración del paquete
// Definimos cada una de las librerias, entre las principales: encoding/json y bytes para el uso de arreglos
// fmt permite mostrar el texto
// html/template reemplaza valores dinámicamente
// io permite leer flujo de datos
// log muestra mensajes en consola
// multipart sirve para procesar formularios con archivos
import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"miweb/constants"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// Representación del endPoint de prueba
type RespuestaCaracteristicas struct {
	Status                   bool              `json:"status"`
	Registros                int               `json:"registros"`
	CaracteristicasJuecesPaz []Caracteristicas `json:"caracteristicas_jueces_paz"`
}

type Caracteristicas struct {
	Id             string `json:"id"`
	Caracteristica string `json:"caracteristica"`
}

// Representación de los elementos recibidos por webLock
type Opcion struct {
	IDOpcion      string `json:"IdOpcion"`
	IDOpcionPadre string `json:"IdOpcionPadre"`
	Nivel         string `json:"Nivel"`
	NombreObjeto  string `json:"NombreObjeto"`
	Descripcion   string `json:"Descripcion"`
	ClaveOpcion   string `json:"ClaveOpcion"`
	URL           string `json:"url"`
	NombreWeb     string `json:"NombreWeb"`
	NewWinWeb     string `json:"NewWinWeb"`
	IDTipoOpcion  string `json:"IdTipoOpcion"`
	Permiso       string `json:"Permiso"`
}

type Usuario struct {
	Token               string   `json:"token"`
	Username            string   `json:"username"`
	UsuarioNT           string   `json:"usuariont"`
	IDUsuario           string   `json:"IdUsuario"`
	TipoUsuario         string   `json:"TipoUsuario"`
	Credencial          string   `json:"Credencial"`
	Nombre              string   `json:"Nombre"`
	ApPaterno           string   `json:"ApPaterno"`
	ApMaterno           string   `json:"ApMaterno"`
	IDDependencia       string   `json:"IdDependencia"`
	Dependencia         string   `json:"Dependencia"`
	ClvDependencia      string   `json:"clvDependencia"`
	IDEdificio          string   `json:"IdEdificio"`
	ClvEdificio         string   `json:"clvEdificio"`
	ClvSistema          string   `json:"clvSistema"`
	Titulo              string   `json:"Titulo"`
	URL                 string   `json:"URL"`
	Icono               string   `json:"Icono"`
	IDRol               string   `json:"IdRol"`
	Usuario             string   `json:"Usuario"`
	RutaRetorno         string   `json:"RutaRetorno"`
	Reporteador         string   `json:"Reporteador"`
	IDPresupuesto       string   `json:"IdPresupuesto"`
	Descripcion         string   `json:"Descripcion"`
	IDSistema           string   `json:"IdSistema"`
	InstitucionRol      string   `json:"InstitucionRol"`
	ClaveDependenciaRol string   `json:"ClaveDependenciaRol"`
	Opciones            []Opcion `json:"Opciones"`
}

// Definimos la siguiente variable para obtener los datos mediante Get una vez que la sesión se encuentre creada
// var sessionStore = map[string]Usuario{}

type Pagina struct {
	Caracteristicas []Caracteristicas
	Usuarios        []Usuario
}

func inicio(w http.ResponseWriter, r *http.Request) {
	log.Println("Método recibido: ", r.Method)
	// Inicializa la lista vacía
	usuarios := []Usuario{}

	// Intentar recuperar usuario desde la cookie cuando el método sea Get
	if r.Method == http.MethodGet {
		log.Println("=== VALIDANDO SESIÓN GET ===")
		cookie, err := r.Cookie("session_id")
		if err != nil {
			log.Println("No existe la cookie de sesión")
			http.Redirect(
				w,
				r,
				constants.InicioSesion,
				http.StatusFound)
			return
		}

		usuario, err := obtenerSesionBFF(cookie.Value)
		if err != nil {
			http.Redirect(
				w,
				r,
				constants.InicioSesion,
				http.StatusFound,
			)
			return
		}

		usuarios = append(usuarios, *usuario)
		log.Println("Cookie encontrada: ", cookie.Value)

		/*if !validarSesionBFF(cookie.Value) {
			log.Println("La sesión ya no existe en el BFF")
			http.Redirect(
				w,
				r,
				constants.InicioSesion,
				http.StatusFound,
			)
			return
		}*/

		log.Println("Sesión válida en el BFF")

		/*usuario, existe := sessionStore[cookie.Value]
		if !existe {
			log.Println("No existe sesión para: ", cookie.Value)
			http.Redirect(w, r, constants.InicioSesion, http.StatusFound)
			return
		}
		log.Println("Usuario recuperado desde sesión: ", usuario.Usuario)
		usuarios = append(usuarios, usuario)*/
	}

	// Validando la existencia de la cookie
	if r.Method == http.MethodPost {

		/*if cookie, err := r.Cookie("session_id"); err == nil {
			if _, existe := sessionStore[cookie.Value]; existe {
				log.Println("Sesión ya existe")
				http.Redirect(
					w,
					r,
					"/",
					http.StatusFound,
				)
				return
			}
		}*/

		log.Println("=== VALIDANDO SESIÓN POST ===")

		cookie, err := r.Cookie("session_id")
		if err == nil {
			log.Println("Cookie encontrada")
			log.Println("SessionId: ", cookie.Value)
		} else {
			log.Println("No existe la cookie session_id")
		}
	}

	caracteristicas, err := obtenerPermisos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Procesar el Post
	if r.Method == http.MethodPost {
		// Leo el contenido
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, fmt.Sprintf("No se pudo leer la solicitud: %v", err), http.StatusBadRequest)
			return
		}

		body = bytes.TrimSpace(body)
		if len(body) == 0 {
			http.Error(w, "El cuerpo de la solicitud está vacío", http.StatusBadRequest)
			return
		}

		contentType := ""
		if header := r.Header.Get("Content-Type"); header != "" {
			var err error
			contentType, _, err = mime.ParseMediaType(header)
			if err != nil {
				http.Error(w, "Encabezado Content-Type inválido", http.StatusBadRequest)
				return
			}
		}
		if contentType == "application/x-www-form-urlencoded" {
			form, err := url.ParseQuery(string(body))
			if err != nil {
				http.Error(w, "No se pudo leer el formulario: "+err.Error(), http.StatusBadRequest)
				return
			}
			jsonValue := form.Get("json")
			if jsonValue == "" {
				jsonValue = form.Get("Json")
			}
			if jsonValue == "" {
				http.Error(w, `El formulario debe incluir un campo "json" o "Json" con JSON válido`, http.StatusBadRequest)
				return
			}
			body = []byte(jsonValue)
		} else if contentType == "multipart/form-data" {
			_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil {
				http.Error(w, "Encabezado Content-Type inválido", http.StatusBadRequest)
				return
			}
			boundary := params["boundary"]
			if boundary == "" {
				http.Error(w, "Falta el boundary del formulario multipart", http.StatusBadRequest)
				return
			}
			form, err := multipart.NewReader(bytes.NewReader(body), boundary).ReadForm(1 << 20)
			if err != nil {
				http.Error(w, "No se pudo leer el formulario multipart: "+err.Error(), http.StatusBadRequest)
				return
			}
			defer form.RemoveAll()
			values := form.Value["json"]
			if len(values) == 0 {
				values = form.Value["Json"]
			}
			if len(values) == 0 {
				http.Error(w, `El formulario debe incluir un campo "json" o "Json"`, http.StatusBadRequest)
				return
			}
			body = []byte(values[0])
		}
		body = bytes.TrimSpace(body)
		if len(body) == 0 {
			http.Error(w, "El campo JSON está vacío", http.StatusBadRequest)
			return
		}

		// Deserializo el json en conjunto con las opciones de seguridad
		usuarios, err = decodeUsuarios(body, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("No se pudieron leer los datos: %v", err), http.StatusBadRequest)
			return
		}
		if len(usuarios) > 0 {
			err = validarUsuario(usuarios[0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		// Compruebo que la lista no se encuentra vacía para poder acceder al campo
		if len(usuarios) > 0 {
			log.Println("Entrando al método del token")
			log.Println("RutaRetorno:", usuarios[0].RutaRetorno)
			token, cookies, err := obtenerPermisosAplicacion(
				usuarios[0].Token,
				usuarios[0].Usuario,
				usuarios[0].Username,
				usuarios[0].UsuarioNT,
				usuarios[0].IDUsuario,
				usuarios[0].TipoUsuario,
				usuarios[0].Credencial,
				usuarios[0].Nombre,
				usuarios[0].ApPaterno,
				usuarios[0].ApMaterno,
				usuarios[0].IDRol,
				usuarios[0].ClvDependencia,
				usuarios[0].ClvSistema,
				usuarios[0].IDSistema,
				usuarios[0].Descripcion,
				usuarios[0].IDDependencia,
				usuarios[0].Titulo,
				usuarios[0].RutaRetorno,
				usuarios[0].Reporteador,
				usuarios[0].IDEdificio,
				usuarios[0].ClvEdificio,
				usuarios[0].Dependencia,
				usuarios[0].IDPresupuesto,
				usuarios[0].InstitucionRol,
				usuarios[0].ClaveDependenciaRol,
				usuarios[0].Opciones,
			)
			if err != nil {
				log.Println("Error obteniendo token: ", err)
				return
			}
			log.Println("El valor del token es: ", token)

			for _, cookie := range cookies {
				// sessionStore[cookie.Value] = usuarios[0]
				log.Printf("Enviando cookie al navegador: %s=%s",
					cookie.Name,
					cookie.Value)
				cookie.Domain = ""
				cookie.Path = "/"
				http.SetCookie(w, cookie)
			}
			http.Redirect(
				w,
				r,
				"/",
				http.StatusFound,
			)
			return
		}
	}

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	datos := Pagina{
		Caracteristicas: caracteristicas,
		Usuarios:        usuarios,
	}

	tmpl, err := template.ParseFiles("templates/index.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Usuarios: ", len(datos.Usuarios))
	log.Println("Características: ", len(datos.Caracteristicas))
	err = tmpl.Execute(w, datos)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func obtenerSesionBFF(sessionID string) (*Usuario, error) {
	log.Println("Entrando a obtenerSesionBFF")
	req, err := http.NewRequest(
		"GET",
		constants.AuthPermissions,
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	})

	client := &http.Client{}

	log.Println("Llamando: ", constants.AuthSession)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	log.Println("Status auth/session: ", resp.Status)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sesión inválida")
	}

	var usuario Usuario

	err = json.NewDecoder(resp.Body).Decode(&usuario)

	if err != nil {
		return nil, err
	}

	return &usuario, nil
}

func validarUsuario(u Usuario) error {
	if u.ApMaterno == "" {
		return fmt.Errorf("ApMaterno no recibido")
	}
	if u.Token == "" {
		return fmt.Errorf("Token invalido")
	}
	log.Println("El ApMaterno es: ", u.ApMaterno)
	log.Println("El valor del Token es: ", u.Token)
	return nil
}

func obtenerPermisosAplicacion(token, usuario, username, usuariont, idUsuario, tipoUsuario, credencial, nombre, apePaterno, apeMaterno, IdRol, claveDependencia, claveSistema, idSistema, descripcion, idDependencia, titulo, rutaretorno, reporteador, idedificio, clvedificio, dependencia, idpresupuesto, institucionrol, clavedependenciarol string, opciones []Opcion) (string, []*http.Cookie, error) {
	payload := map[string]interface{}{
		"Usuario":             usuario,
		"username":            username,
		"UsuarioNT":           usuariont,
		"IdUsuario":           idUsuario,
		"TipoUsuario":         tipoUsuario,
		"Credencial":          credencial,
		"Nombre":              nombre,
		"ApPaterno":           apePaterno,
		"ApMaterno":           apeMaterno,
		"IdRol":               IdRol,
		"clvDependencia":      claveDependencia,
		"clvSistema":          claveSistema,
		"IdSistema":           idSistema,
		"token":               token,
		"Descripcion":         descripcion,
		"IdDependencia":       idDependencia,
		"Titulo":              titulo,
		"RutaRetorno":         rutaretorno,
		"Reporteador":         reporteador,
		"IdEdificio":          idedificio,
		"clvEdificio":         clvedificio,
		"Dependencia":         dependencia,
		"IdPresupuesto":       idpresupuesto,
		"InstitucionRol":      institucionrol,
		"ClaveDependenciaRol": clavedependenciarol,
		"Opciones":            opciones,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}

	log.Printf("Cantidad de opciones enviadas: %d", len(opciones))
	for _, op := range opciones {
		log.Printf(
			"Opción %s - %s",
			op.IDOpcion,
			op.Descripcion,
		)
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Construyo el json
	form := url.Values{}
	form.Set("json", string(jsonData))

	log.Println("Iniciando el envío de valor al token")
	// La siguiente línea es para mostrar el json que se obtiene del WebLog
	// log.Println(string(jsonData))
	resp, err := client.PostForm(
		//"https://localhost:7096/auth/callback",
		constants.AuthCallback,
		form,
	)

	log.Println("Status:", resp.Status)
	for k, v := range resp.Header {
		log.Printf("%s = %v", k, v)
	}

	if err != nil {
		return "", nil, err
	}

	// Comprobación de Cookies
	log.Println("Comprobación de cookies")
	log.Printf("Cantidad de cookies: %d", len(resp.Cookies()))
	for k, v := range resp.Header {
		log.Printf("%s: %v", k, v)
	}

	for _, cookie := range resp.Cookies() {
		log.Printf("Cookie: %s=%s",
			cookie.Name,
			cookie.Value)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	log.Println("Código de respuesta es:", resp.Status)
	log.Println("URL final es: ", resp.Request.URL)
	return string(body), resp.Cookies(), nil
}

/*func validarSesionBFF(sessionID string) bool {
	req, err := http.NewRequest(
		"GET",
		constants.AuthCheck,
		nil,
	)
	if err != nil {
		return false
	}

	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	})
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Error llamado auth/check:", err)
		return false
	}
	defer resp.Body.Close()
	log.Println("Status auth/check:", resp.Status)
	return resp.StatusCode == http.StatusOK
}*/

// Se define la lista de estructura
func decodeUsuarios(data []byte, depth int) ([]Usuario, error) {
	if depth > 3 {
		return nil, fmt.Errorf("formato de datos anidado demasiadas veces")
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("el campo json está vacío")
	}

	var usuarios []Usuario
	if data[0] == '[' {
		if err := json.Unmarshal(data, &usuarios); err == nil {
			return usuarios, nil
		}
	}

	if data[0] == '{' {
		var objeto map[string]json.RawMessage
		if err := json.Unmarshal(data, &objeto); err == nil {
			for key, value := range objeto {
				if strings.EqualFold(key, "json") {
					return decodeUsuarios(value, depth+1)
				}
			}
			var usuario Usuario
			if err := json.Unmarshal(data, &usuario); err == nil {
				return []Usuario{usuario}, nil
			}
		}
	}

	var texto string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &texto); err == nil {
			return decodeUsuarios([]byte(texto), depth+1)
		}
	}

	normalizado, err := singleQuotedToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("JSON inválido: %w", err)
	}
	if bytes.Equal(normalizado, data) {
		return nil, fmt.Errorf("se esperaba un arreglo, un objeto o un campo json con datos")
	}
	return decodeUsuarios(normalizado, depth+1)
}

func singleQuotedToJSON(data []byte) ([]byte, error) {
	var output bytes.Buffer
	var quote byte
	escaped := false

	for i := 0; i < len(data); i++ {
		char := data[i]
		if quote == 0 {
			if char == '\'' {
				quote = char
				output.WriteByte('"')
			} else {
				output.WriteByte(char)
				if char == '"' {
					quote = char
				}
			}
			continue
		}

		if escaped {
			switch char {
			case '\'':
				if quote == '\'' {
					output.WriteByte('\'')
				} else {
					output.WriteByte('\\')
					output.WriteByte(char)
				}
			case '"':
				if quote == '\'' {
					output.WriteString(`\"`)
				} else {
					output.WriteByte('\\')
					output.WriteByte(char)
				}
			case '\\', '/', 'b', 'f', 'n', 'r', 't':
				output.WriteByte('\\')
				output.WriteByte(char)
			case 'u':
				if i+4 >= len(data) {
					return nil, fmt.Errorf("escape Unicode incompleto")
				}
				output.WriteString(`\u`)
				for j := 1; j <= 4; j++ {
					output.WriteByte(data[i+j])
				}
				i += 4
			default:
				return nil, fmt.Errorf("secuencia de escape no válida: \\%c", char)
			}
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
		} else if char == quote {
			quote = 0
			output.WriteByte('"')
		} else if char == '"' && quote == '\'' {
			output.WriteString(`\"`)
		} else if char < 0x20 {
			return nil, fmt.Errorf("carácter de control dentro de una cadena")
		} else {
			output.WriteByte(char)
		}
	}

	if escaped || quote != 0 {
		return nil, fmt.Errorf("cadena sin cerrar o escape incompleto")
	}
	return output.Bytes(), nil
}

func obtenerPermisos() ([]Caracteristicas, error) {
	// En este punto se llama al BFF
	log.Println("Iniciando llamada al BFF...")
	resp, err := http.Get(
		//"http://localhost:5165/api/PermisosUsuario",
		//"https://localhost:7096/api/PermisosUsuario",
		constants.PermisosUsuario,
	)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	defer resp.Body.Close()

	log.Println("Código respuesta:", resp.Status)

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("Error leyendo respuesta:", err)
		return nil, err
	}

	var respuesta RespuestaCaracteristicas

	err = json.Unmarshal(body, &respuesta)

	if err != nil {
		return nil, err
	}

	/*log.Println("Resultado BFF:")
	log.Println(string(body))*/

	return respuesta.CaracteristicasJuecesPaz, nil
}

func VerCookies(w http.ResponseWriter, r *http.Request) {

	for _, c := range r.Cookies() {
		fmt.Fprintf(
			w,
			"%s = %s\n",
			c.Name,
			c.Value,
		)
	}
}

func InicioSesion(w http.ResponseWriter, r *http.Request) {
	http.Redirect(
		w,
		r,
		constants.InicioSesion,
		http.StatusFound,
	)
}

func InicioHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Cargando de nuevo la aplicación")
	log.Println("Método recibido: ", r.Method)
	http.ServeFile(w, r, "index.html")
}

// Función que se ejecuta al momento de cargar
func main() {
	log.Println("Servidor iniciado")
	http.HandleFunc("/cookies", VerCookies)
	http.HandleFunc("/iniciosesion", InicioSesion)
	http.HandleFunc("/", inicio)
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
