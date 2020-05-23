package main

import (
  "os"
  "fmt"
  "encoding/base64"
  "encoding/asn1"
  "crypto/sha256"
  "crypto/ecdsa"
  "crypto/x509"
  "io/ioutil"
  "math/big"
  "strings"
  "strconv"
  "github.com/gin-gonic/gin"
  "github.com/joho/godotenv"
)

func main() {
  err_read := godotenv.Load()
  if err_read != nil {
    fmt.Printf("error: %v", err_read)
  }

	r := gin.Default()
  r.POST("/", func(c *gin.Context) {

    // 1. Get the signature from the "X-Twilio-Email-Event-Webhook-Signature" HTTP header.
    HEADER_SIGNATURE := "X-Twilio-Email-Event-Webhook-Signature"
    s := c.GetHeader(HEADER_SIGNATURE)
    fmt.Printf("%s: %s\n", HEADER_SIGNATURE, s)

    // 2. Get the timestamp from the ""X-Twilio-Email-Event-Webhook-Signature" HTTP header.
    HEADER_TIMESTAMP := "X-Twilio-Email-Event-Webhook-Timestamp"
    ts := c.GetHeader(HEADER_TIMESTAMP)
    fmt.Printf("%s: %s\n", HEADER_TIMESTAMP, ts)

    // Payload
    payload, _ := ioutil.ReadAll(c.Request.Body)
    fmt.Printf("payload: %s\n", payload)

    // 3. Base64 decode the signature. Then perform an ASN.1 unmarshal of the decoded signature into a string. This string will be in the form of {r value},{s value}.
    signatureBytes, _ := base64.StdEncoding.DecodeString(s)
    fmt.Printf("signatureBytes: %s\n", bytes2Str(signatureBytes))
    type rs struct {
      R *big.Int
      S *big.Int
    }
    ecdsaSig := rs{}
    asn1.Unmarshal(signatureBytes, &ecdsaSig)

    // 4. Generate a sha256 hash of the timestamp + payload (use raw bytes).
    tsBytes := []byte(ts)
    h := sha256.New()
    h.Write(tsBytes)
    h.Write(payload)
    hashedPayload := h.Sum(nil)
    fmt.Printf("hashedPayload: %s\n", bytes2Str(hashedPayload))

    // 4.1. Make PublicKey
    SG_VERIFICATION_KEY := os.Getenv("SG_VERIFICATION_KEY")
    keyBytes, _ := base64.StdEncoding.DecodeString(SG_VERIFICATION_KEY)
    fmt.Printf("keyBytes: %s\n", bytes2Str(keyBytes))
    pub, _ := x509.ParsePKIXPublicKey(keyBytes)

    // 5. Verify the signature.
    result := ecdsa.Verify(pub.(*ecdsa.PublicKey), hashedPayload, ecdsaSig.R, ecdsaSig.S)
    if (!result) {
      panic(result);
    } else {
      fmt.Printf("result: %s\n", strconv.FormatBool(result))
      c.JSON(200, gin.H{
        "message": "pong",
      })
    }
  })
	r.Run(":3000") // listen and serve on 0.0.0.0:3000
}

func bytes2Str(bytes ...[]byte) string {
    strs := []string{}
    for _, b := range bytes {
        strs = append(strs, fmt.Sprintf("%02x", b))
    }
    return strings.Join(strs, " ")
}
