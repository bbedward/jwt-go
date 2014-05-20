// A useful example app.  You can use this to debug your tokens on the command line.
// This is also a great place to look at how you might use this library.
//
// Example usage:
// The following will create and sign a token, then verify it and output the original claims.
//     echo {\"foo\":\"bar\"} | bin/jwt -key test/sample_key -alg RS256 -sign - | bin/jwt -key test/sample_key.pub -verify -
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/dgrijalva/jwt-go"
)

var (
	// Options
	flagAlg    = flag.String("alg", "", "signing algorithm identifier")
	flagKey    = flag.String("key", "", "path to key file or '-' to read from stdin")
	flagPretty = flag.Bool("pretty", true, "output pretty JSON")

	// Modes - exactly one of these is required
	flagSign   = flag.String("sign", "", "path to claims object to sign or '-' to read from stdin")
	flagVerify = flag.String("verify", "", "path to JWT token to verify or '-' to read from stdin")
)

func main() {
	// Usage message if you ask for -help or if you mess up inputs.
	// TODO: make this better
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  One of the following flags is required: sign, verify\n")
		flag.PrintDefaults()
	}

	// Parse command line options
	flag.Parse()

	// Do the thing.  If something goes wrong, print error to stderr
	// and exit with a non-zero status code
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Figure out which thing to do and then do that
func start() error {
	if *flagSign != "" {
		return signToken()
	} else if *flagVerify != "" {
		return verifyToken()
	} else {
		flag.Usage()
		return fmt.Errorf("None of the required flags are present.  What do you want me to do?")
	}
}

// Helper func:  Read input from specified file or stdin
func loadData(p string) ([]byte, error) {
	if p == "" {
		return nil, fmt.Errorf("No path specified")
	}

	var rdr io.Reader
	if p == "-" {
		rdr = os.Stdin
	} else {
		if f, err := os.Open(p); err == nil {
			rdr = f
			defer f.Close()
		} else {
			return nil, err
		}
	}
	return ioutil.ReadAll(rdr)
}

// Verify a token and output the claims
func verifyToken() error {
	// get the token
	tokData, err := loadData(*flagVerify)
	if err != nil {
		return fmt.Errorf("Couldn't read token: %v", err)
	}

	// Parse the token.  Load the key from command line option
	token, err := jwt.Parse(string(tokData), func(t *jwt.Token) ([]byte, error) {
		return loadData(*flagKey)
	})

	// Print an error if we can't parse for some reason
	if err != nil {
		return err
	}

	// Is token invalid?
	if !token.Valid {
		return fmt.Errorf("Token is invalid")
	}

	// Print the token details
	// TODO: observe the pretty flag
	if out, err := json.MarshalIndent(token.Claims, "", "    "); err == nil {
		fmt.Println(string(out))
	} else {
		return fmt.Errorf("Failed to output claims: %v", err)
	}

	return nil
}

// Create, sign, and output a token
func signToken() error {
	// get the token
	tokData, err := loadData(*flagSign)
	if err != nil {
		return fmt.Errorf("Couldn't read token: %v", err)
	}

	// parse the JSON of the claims
	var claims map[string]interface{}
	if err := json.Unmarshal(tokData, &claims); err != nil {
		return fmt.Errorf("Couldn't parse claims JSON: %v", err)
	}

	// get the key
	keyData, err := loadData(*flagKey)
	if err != nil {
		return fmt.Errorf("Couldn't read key: %v", err)
	}

	// get the signing alg
	alg := jwt.GetSigningMethod(*flagAlg)
	if alg == nil {
		return fmt.Errorf("Couldn't find signing method: %v", *flagAlg)
	}

	// create a new token
	token := jwt.New(alg)
	token.Claims = claims

	if out, err := token.SignedString(keyData); err == nil {
		fmt.Println(out)
	} else {
		return fmt.Errorf("Error signing token: %v", err)
	}

	return nil
}
