package httpr

const (
	FormatCsv  = "csv"
	FormatJson = "json"
	FormatXml  = "xml"

	CompressionGzip    = "gzip"
	CompressionTar     = "tar"
	CompressionDeflate = "deflate"
	CompressionNone    = ""

	AuthBasic  = "basic"
	AuthBearer = "bearer"
	AuthOAuth2 = "oauth2"
	AuthNone   = "none"
)
