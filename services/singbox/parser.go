package singbox

import (
	"net/url"

	"github.com/FMotalleb/go-tools/builder"
	"github.com/spf13/cast"
)

// TryParseOutboundURL constructs a Singbox outbound configuration map from the provided URL and its query parameters.
//
// The function extracts relevant fields such as type, server, port, UUID, transport settings, and TLS parameters from the URL and its query string. The resulting map is structured under the key path "core.singbox.outbounds" and is suitable for use in Singbox configuration files.
//
// Returns the constructed configuration map and a nil error.
func TryParseOutboundURL(url *url.URL) (map[string]any, error) {
	cb := builder.NewNested()

	query := url.Query()

	cb.Set("type", url.Scheme).
		Set("tag", "proxy").
		Set("packet_encoding", query.Get("packetEncoding")).
		Set("server", url.Hostname())
	port := url.Port()
	if port == "" {
		port = "443"
	}
	cb.Set("server_port", cast.ToUint16(port))

	if url.User != nil {
		cb.Set("uuid", url.User.Username())
	}
	loadTLSParams(cb, query)

	cb.Set("transport.type", query.Get("type"))
	switch query.Get("type") {
	case "tcp", "ws", "http", "httpupgrade":
		cb.Set("transport.path", query.Get("path"))
		cb.Set("transport.headers.Host", query.Get("host"))
	}
	sn := query.Get("serviceName")
	if sn != "" {
		cb.Set("transport.service_name", query.Get("serviceName"))
	}
	cb.Set("flow", query.Get("flow"))
	out := cb.Data

	return map[string]any{
		"core": map[string]any{
			"singbox": map[string]any{
				"outbounds": []map[string]any{
					out,
				},
			},
		},
	}, nil
}

// loadTLSParams adds TLS-related configuration fields to the builder based on URL query parameters.
// It enables and configures TLS settings if the "security" parameter is set to "tls", including options for insecure connections, server name indication, uTLS fingerprint, and Reality public key and short ID.
func loadTLSParams(cb *builder.Nested, query url.Values) {
	if query.Get("security") != "tls" {
		return
	}
	cb.Set("tls.enabled", true)

	insec := query.Get("allowInsecure")

	if insec == "" {
		insec = "0"
	}
	cb.Set("tls.insecure", cast.ToBool(insec)).
		Set("tls.server_name", query.Get("sni"))

	if query.Get("fp") != "" {
		cb.Set("tls.utls.enabled", true).
			Set("tls.utls.fingerprint", query.Get("fp"))
	}
	if query.Get("pbk") != "" {
		cb.Set("tls.reality.enabled", true).
			Set("tls.reality.public_key", query.Get("pbk")).
			Set("tls.reality.short_id", query.Get("sid"))
	}
}
