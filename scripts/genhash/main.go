// Genera una contraseña fuerte al azar y su hash bcrypt, para rotar
// credenciales sin que nada quede escrito en el repositorio.
//
//	go run ./scripts/genhash              → genera una contraseña nueva
//	go run ./scripts/genhash "mi-clave"   → solo hashea la que le pases
//
// La salida va a stdout: cópiala, úsala y no la guardes en ningún archivo
// versionado.
package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%*-_"

func randomPassword(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[idx.Int64()]
	}
	return string(b), nil
}

func main() {
	password := ""
	if len(os.Args) > 1 {
		password = os.Args[1]
	} else {
		p, err := randomPassword(20)
		if err != nil {
			fmt.Fprintln(os.Stderr, "no se pudo generar la contraseña:", err)
			os.Exit(1)
		}
		password = p
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "no se pudo generar el hash:", err)
		os.Exit(1)
	}

	fmt.Printf("Contraseña : %s\n", password)
	fmt.Printf("Hash bcrypt: %s\n\n", hash)
	fmt.Println("SQL de rotación (ajusta el correo destino):")
	fmt.Printf("  UPDATE users SET password_hash = '%s', must_change_password = true\n", hash)
	fmt.Printf("  WHERE role = 'admin' AND lower(email) = lower('admin@puntofusion.local');\n")
}
