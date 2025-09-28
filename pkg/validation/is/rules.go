// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package is provides a list of commonly used string validation rules.
package is

import (
	"regexp"
	"unicode"

	"github.com/adverax/metacrm/pkg/validation"
	"github.com/asaskevich/govalidator"
)

const (
	RuleTypeEmail            validation.RuleType = "email"
	RuleTypeEmailFormat      validation.RuleType = "email_format"
	RuleTypeURL              validation.RuleType = "url"
	RuleTypeRequestURL       validation.RuleType = "request_url"
	RuleTypeRequestURI       validation.RuleType = "request_uri"
	RuleTypeAlpha            validation.RuleType = "alpha"
	RuleTypeDigit            validation.RuleType = "digit"
	RuleTypeAlphanumeric     validation.RuleType = "alnum"
	RuleTypeUTFLetter        validation.RuleType = "utf_letter"
	RuleTypeUTFDigit         validation.RuleType = "utf_digit"
	RuleTypeUTFLetterNumeric validation.RuleType = "utf_letter_numeric"
	RuleTypeUTFNumeric       validation.RuleType = "utf_numeric"
	RuleTypeLowerCase        validation.RuleType = "lower_case"
	RuleTypeUpperCase        validation.RuleType = "upper_case"
	RuleTypeHexadecimal      validation.RuleType = "hexadecimal"
	RuleTypeHexColor         validation.RuleType = "hex_color"
	RuleTypeRGBColor         validation.RuleType = "rgb_color"
	RuleTypeInt              validation.RuleType = "int"
	RuleTypeFloat            validation.RuleType = "float"
	RuleTypeUUIDv3           validation.RuleType = "uuid_v3"
	RuleTypeUUIDv4           validation.RuleType = "uuid_v4"
	RuleTypeUUIDv5           validation.RuleType = "uuid_v5"
	RuleTypeUUID             validation.RuleType = "uuid"
	RuleTypeCreditCard       validation.RuleType = "credit_card"
	RuleTypeISBN10           validation.RuleType = "isbn_10"
	RuleTypeISBN13           validation.RuleType = "isbn_13"
	RuleTypeISBN             validation.RuleType = "isbn"
	RuleTypeJSON             validation.RuleType = "json"
	RuleTypeASCII            validation.RuleType = "ascii"
	RuleTypePrintableASCII   validation.RuleType = "printable_ascii"
	RuleTypeMultibyte        validation.RuleType = "multibyte"
	RuleTypeFullWidth        validation.RuleType = "full_width"
	RuleTypeHalfWidth        validation.RuleType = "half_width"
	RuleTypeVariableWidth    validation.RuleType = "variable_width"
	RuleTypeBase64           validation.RuleType = "base64"
	RuleTypeDataURI          validation.RuleType = "data_uri"
	RuleTypeE164             validation.RuleType = "e164"
	RuleTypeCountryCode2     validation.RuleType = "country_code_2_letter"
	RuleTypeCountryCode3     validation.RuleType = "country_code_3_letter"
	RuleTypeCurrencyCode     validation.RuleType = "currency_code"
	RuleTypeDialString       validation.RuleType = "dial_string"
	RuleTypeMAC              validation.RuleType = "mac"
	RuleTypeIP               validation.RuleType = "ip"
	RuleTypeIPv4             validation.RuleType = "ipv4"
	RuleTypeIPv6             validation.RuleType = "ipv6"
	RuleTypeSubdomain        validation.RuleType = "sub_domain"
	RuleTypeDomain           validation.RuleType = "domain"
	RuleTypeDNSName          validation.RuleType = "dns_name"
	RuleTypeHost             validation.RuleType = "host"
	RuleTypePort             validation.RuleType = "port"
	RuleTypeMongoID          validation.RuleType = "mongo_id"
	RuleTypeLatitude         validation.RuleType = "latitude"
	RuleTypeLongitude        validation.RuleType = "longitude"
	RuleTypeSSN              validation.RuleType = "ssn"
	RuleTypeSemver           validation.RuleType = "semver"
)

var (
	// ErrEmail is the error that returns in case of an invalid email.
	ErrEmail = validation.NewError("validation_is_email", "must be a valid email address")
	// ErrURL is the error that returns in case of an invalid URL.
	ErrURL = validation.NewError("validation_is_url", "must be a valid URL")
	// ErrRequestURL is the error that returns in case of an invalid request URL.
	ErrRequestURL = validation.NewError("validation_is_request_url", "must be a valid request URL")
	// ErrRequestURI is the error that returns in case of an invalid request URI.
	ErrRequestURI = validation.NewError("validation_request_is_request_uri", "must be a valid request URI")
	// ErrAlpha is the error that returns in case of an invalid alpha value.
	ErrAlpha = validation.NewError("validation_is_alpha", "must contain English letters only")
	// ErrDigit is the error that returns in case of an invalid digit value.
	ErrDigit = validation.NewError("validation_is_digit", "must contain digits only")
	// ErrAlphanumeric is the error that returns in case of an invalid alphanumeric value.
	ErrAlphanumeric = validation.NewError("validation_is_alphanumeric", "must contain English letters and digits only")
	// ErrUTFLetter is the error that returns in case of an invalid utf letter value.
	ErrUTFLetter = validation.NewError("validation_is_utf_letter", "must contain unicode letter characters only")
	// ErrUTFDigit is the error that returns in case of an invalid utf digit value.
	ErrUTFDigit = validation.NewError("validation_is_utf_digit", "must contain unicode decimal digits only")
	// ErrUTFLetterNumeric is the error that returns in case of an invalid utf numeric or letter value.
	ErrUTFLetterNumeric = validation.NewError("validation_is utf_letter_numeric", "must contain unicode letters and numbers only")
	// ErrUTFNumeric is the error that returns in case of an invalid utf numeric value.
	ErrUTFNumeric = validation.NewError("validation_is_utf_numeric", "must contain unicode number characters only")
	// ErrLowerCase is the error that returns in case of an invalid lower case value.
	ErrLowerCase = validation.NewError("validation_is_lower_case", "must be in lower case")
	// ErrUpperCase is the error that returns in case of an invalid upper case value.
	ErrUpperCase = validation.NewError("validation_is_upper_case", "must be in upper case")
	// ErrHexadecimal is the error that returns in case of an invalid hexadecimal number.
	ErrHexadecimal = validation.NewError("validation_is_hexadecimal", "must be a valid hexadecimal number")
	// ErrHexColor is the error that returns in case of an invalid hexadecimal color code.
	ErrHexColor = validation.NewError("validation_is_hex_color", "must be a valid hexadecimal color code")
	// ErrRGBColor is the error that returns in case of an invalid RGB color code.
	ErrRGBColor = validation.NewError("validation_is_rgb_color", "must be a valid RGB color code")
	// ErrInt is the error that returns in case of an invalid integer value.
	ErrInt = validation.NewError("validation_is_int", "must be an integer number")
	// ErrFloat is the error that returns in case of an invalid float value.
	ErrFloat = validation.NewError("validation_is_float", "must be a floating point number")
	// ErrUUIDv3 is the error that returns in case of an invalid UUIDv3 value.
	ErrUUIDv3 = validation.NewError("validation_is_uuid_v3", "must be a valid UUID v3")
	// ErrUUIDv4 is the error that returns in case of an invalid UUIDv4 value.
	ErrUUIDv4 = validation.NewError("validation_is_uuid_v4", "must be a valid UUID v4")
	// ErrUUIDv5 is the error that returns in case of an invalid UUIDv5 value.
	ErrUUIDv5 = validation.NewError("validation_is_uuid_v5", "must be a valid UUID v5")
	// ErrUUID is the error that returns in case of an invalid UUID value.
	ErrUUID = validation.NewError("validation_is_uuid", "must be a valid UUID")
	// ErrCreditCard is the error that returns in case of an invalid credit card number.
	ErrCreditCard = validation.NewError("validation_is_credit_card", "must be a valid credit card number")
	// ErrISBN10 is the error that returns in case of an invalid ISBN-10 value.
	ErrISBN10 = validation.NewError("validation_is_isbn_10", "must be a valid ISBN-10")
	// ErrISBN13 is the error that returns in case of an invalid ISBN-13 value.
	ErrISBN13 = validation.NewError("validation_is_isbn_13", "must be a valid ISBN-13")
	// ErrISBN is the error that returns in case of an invalid ISBN value.
	ErrISBN = validation.NewError("validation_is_isbn", "must be a valid ISBN")
	// ErrJSON is the error that returns in case of an invalid JSON.
	ErrJSON = validation.NewError("validation_is_json", "must be in valid JSON format")
	// ErrASCII is the error that returns in case of an invalid ASCII.
	ErrASCII = validation.NewError("validation_is_ascii", "must contain ASCII characters only")
	// ErrPrintableASCII is the error that returns in case of an invalid printable ASCII value.
	ErrPrintableASCII = validation.NewError("validation_is_printable_ascii", "must contain printable ASCII characters only")
	// ErrMultibyte is the error that returns in case of an invalid multibyte value.
	ErrMultibyte = validation.NewError("validation_is_multibyte", "must contain multibyte characters")
	// ErrFullWidth is the error that returns in case of an invalid full-width value.
	ErrFullWidth = validation.NewError("validation_is_full_width", "must contain full-width characters")
	// ErrHalfWidth is the error that returns in case of an invalid half-width value.
	ErrHalfWidth = validation.NewError("validation_is_half_width", "must contain half-width characters")
	// ErrVariableWidth is the error that returns in case of an invalid variable width value.
	ErrVariableWidth = validation.NewError("validation_is_variable_width", "must contain both full-width and half-width characters")
	// ErrBase64 is the error that returns in case of an invalid base54 value.
	ErrBase64 = validation.NewError("validation_is_base64", "must be encoded in Base64")
	// ErrDataURI is the error that returns in case of an invalid data URI.
	ErrDataURI = validation.NewError("validation_is_data_uri", "must be a Base64-encoded data URI")
	// ErrE164 is the error that returns in case of an invalid e165.
	ErrE164 = validation.NewError("validation_is_e164_number", "must be a valid E164 number")
	// ErrCountryCode2 is the error that returns in case of an invalid two-letter country code.
	ErrCountryCode2 = validation.NewError("validation_is_country_code_2_letter", "must be a valid two-letter country code")
	// ErrCountryCode3 is the error that returns in case of an invalid three-letter country code.
	ErrCountryCode3 = validation.NewError("validation_is_country_code_3_letter", "must be a valid three-letter country code")
	// ErrCurrencyCode is the error that returns in case of an invalid currency code.
	ErrCurrencyCode = validation.NewError("validation_is_currency_code", "must be valid ISO 4217 currency code")
	// ErrDialString is the error that returns in case of an invalid string.
	ErrDialString = validation.NewError("validation_is_dial_string", "must be a valid dial string")
	// ErrMac is the error that returns in case of an invalid mac address.
	ErrMac = validation.NewError("validation_is_mac_address", "must be a valid MAC address")
	// ErrIP is the error that returns in case of an invalid IP.
	ErrIP = validation.NewError("validation_is_ip", "must be a valid IP address")
	// ErrIPv4 is the error that returns in case of an invalid IPv4.
	ErrIPv4 = validation.NewError("validation_is_ipv4", "must be a valid IPv4 address")
	// ErrIPv6 is the error that returns in case of an invalid IPv6.
	ErrIPv6 = validation.NewError("validation_is_ipv6", "must be a valid IPv6 address")
	// ErrSubdomain is the error that returns in case of an invalid subdomain.
	ErrSubdomain = validation.NewError("validation_is_sub_domain", "must be a valid subdomain")
	// ErrDomain is the error that returns in case of an invalid domain.
	ErrDomain = validation.NewError("validation_is_domain", "must be a valid domain")
	// ErrDNSName is the error that returns in case of an invalid DNS name.
	ErrDNSName = validation.NewError("validation_is_dns_name", "must be a valid DNS name")
	// ErrHost is the error that returns in case of an invalid host.
	ErrHost = validation.NewError("validation_is_host", "must be a valid IP address or DNS name")
	// ErrPort is the error that returns in case of an invalid port.
	ErrPort = validation.NewError("validation_is_port", "must be a valid port number")
	// ErrMongoID is the error that returns in case of an invalid MongoID.
	ErrMongoID = validation.NewError("validation_is_mongo_id", "must be a valid hex-encoded MongoDB ObjectId")
	// ErrLatitude is the error that returns in case of an invalid latitude.
	ErrLatitude = validation.NewError("validation_is_latitude", "must be a valid latitude")
	// ErrLongitude is the error that returns in case of an invalid longitude.
	ErrLongitude = validation.NewError("validation_is_longitude", "must be a valid longitude")
	// ErrSSN is the error that returns in case of an invalid SSN.
	ErrSSN = validation.NewError("validation_is_ssn", "must be a valid social security number")
	// ErrSemver is the error that returns in case of an invalid semver.
	ErrSemver = validation.NewError("validation_is_semver", "must be a valid semantic version")
)

var (
	// Email validates if a string is an email or not. It also checks if the MX record exists for the email domain.
	Email = validation.NewStringRuleWithError(RuleTypeEmail, govalidator.IsExistingEmail, ErrEmail)
	// EmailFormat validates if a string is an email or not. Note that it does NOT check if the MX record exists or not.
	EmailFormat = validation.NewStringRuleWithError(RuleTypeEmailFormat, govalidator.IsEmail, ErrEmail)
	// URL validates if a string is a valid URL
	URL = validation.NewStringRuleWithError(RuleTypeURL, govalidator.IsURL, ErrURL)
	// RequestURL validates if a string is a valid request URL
	RequestURL = validation.NewStringRuleWithError(RuleTypeRequestURL, govalidator.IsRequestURL, ErrRequestURL)
	// RequestURI validates if a string is a valid request URI
	RequestURI = validation.NewStringRuleWithError(RuleTypeRequestURI, govalidator.IsRequestURI, ErrRequestURI)
	// Alpha validates if a string contains English letters only (a-zA-Z)
	Alpha = validation.NewStringRuleWithError(RuleTypeAlpha, govalidator.IsAlpha, ErrAlpha)
	// Digit validates if a string contains digits only (0-9)
	Digit = validation.NewStringRuleWithError(RuleTypeDigit, isDigit, ErrDigit)
	// Alphanumeric validates if a string contains English letters and digits only (a-zA-Z0-9)
	Alphanumeric = validation.NewStringRuleWithError(RuleTypeAlphanumeric, govalidator.IsAlphanumeric, ErrAlphanumeric)
	// UTFLetter validates if a string contains unicode letters only
	UTFLetter = validation.NewStringRuleWithError(RuleTypeUTFLetter, govalidator.IsUTFLetter, ErrUTFLetter)
	// UTFDigit validates if a string contains unicode decimal digits only
	UTFDigit = validation.NewStringRuleWithError(RuleTypeUTFDigit, govalidator.IsUTFDigit, ErrUTFDigit)
	// UTFLetterNumeric validates if a string contains unicode letters and numbers only
	UTFLetterNumeric = validation.NewStringRuleWithError(RuleTypeUTFLetterNumeric, govalidator.IsUTFLetterNumeric, ErrUTFLetterNumeric)
	// UTFNumeric validates if a string contains unicode number characters (category N) only
	UTFNumeric = validation.NewStringRuleWithError(RuleTypeUTFNumeric, isUTFNumeric, ErrUTFNumeric)
	// LowerCase validates if a string contains lower case unicode letters only
	LowerCase = validation.NewStringRuleWithError(RuleTypeLowerCase, govalidator.IsLowerCase, ErrLowerCase)
	// UpperCase validates if a string contains upper case unicode letters only
	UpperCase = validation.NewStringRuleWithError(RuleTypeUpperCase, govalidator.IsUpperCase, ErrUpperCase)
	// Hexadecimal validates if a string is a valid hexadecimal number
	Hexadecimal = validation.NewStringRuleWithError(RuleTypeHexadecimal, govalidator.IsHexadecimal, ErrHexadecimal)
	// HexColor validates if a string is a valid hexadecimal color code
	HexColor = validation.NewStringRuleWithError(RuleTypeHexColor, govalidator.IsHexcolor, ErrHexColor)
	// RGBColor validates if a string is a valid RGB color in the form of rgb(R, G, B)
	RGBColor = validation.NewStringRuleWithError(RuleTypeRGBColor, govalidator.IsRGBcolor, ErrRGBColor)
	// Int validates if a string is a valid integer number
	Int = validation.NewStringRuleWithError(RuleTypeInt, govalidator.IsInt, ErrInt)
	// Float validates if a string is a floating point number
	Float = validation.NewStringRuleWithError(RuleTypeFloat, govalidator.IsFloat, ErrFloat)
	// UUIDv3 validates if a string is a valid version 3 UUID
	UUIDv3 = validation.NewStringRuleWithError(RuleTypeUUIDv3, govalidator.IsUUIDv3, ErrUUIDv3)
	// UUIDv4 validates if a string is a valid version 4 UUID
	UUIDv4 = validation.NewStringRuleWithError(RuleTypeUUIDv4, govalidator.IsUUIDv4, ErrUUIDv4)
	// UUIDv5 validates if a string is a valid version 5 UUID
	UUIDv5 = validation.NewStringRuleWithError(RuleTypeUUIDv5, govalidator.IsUUIDv5, ErrUUIDv5)
	// UUID validates if a string is a valid UUID
	UUID = validation.NewStringRuleWithError(RuleTypeUUID, govalidator.IsUUID, ErrUUID)
	// CreditCard validates if a string is a valid credit card number
	CreditCard = validation.NewStringRuleWithError(RuleTypeCreditCard, govalidator.IsCreditCard, ErrCreditCard)
	// ISBN10 validates if a string is an ISBN version 10
	ISBN10 = validation.NewStringRuleWithError(RuleTypeISBN10, govalidator.IsISBN10, ErrISBN10)
	// ISBN13 validates if a string is an ISBN version 13
	ISBN13 = validation.NewStringRuleWithError(RuleTypeISBN13, govalidator.IsISBN13, ErrISBN13)
	// ISBN validates if a string is an ISBN (either version 10 or 13)
	ISBN = validation.NewStringRuleWithError(RuleTypeISBN, isISBN, ErrISBN)
	// JSON validates if a string is in valid JSON format
	JSON = validation.NewStringRuleWithError(RuleTypeJSON, govalidator.IsJSON, ErrJSON)
	// ASCII validates if a string contains ASCII characters only
	ASCII = validation.NewStringRuleWithError(RuleTypeASCII, govalidator.IsASCII, ErrASCII)
	// PrintableASCII validates if a string contains printable ASCII characters only
	PrintableASCII = validation.NewStringRuleWithError(RuleTypePrintableASCII, govalidator.IsPrintableASCII, ErrPrintableASCII)
	// Multibyte validates if a string contains multibyte characters
	Multibyte = validation.NewStringRuleWithError(RuleTypeMultibyte, govalidator.IsMultibyte, ErrMultibyte)
	// FullWidth validates if a string contains full-width characters
	FullWidth = validation.NewStringRuleWithError(RuleTypeFullWidth, govalidator.IsFullWidth, ErrFullWidth)
	// HalfWidth validates if a string contains half-width characters
	HalfWidth = validation.NewStringRuleWithError(RuleTypeHalfWidth, govalidator.IsHalfWidth, ErrHalfWidth)
	// VariableWidth validates if a string contains both full-width and half-width characters
	VariableWidth = validation.NewStringRuleWithError(RuleTypeVariableWidth, govalidator.IsVariableWidth, ErrVariableWidth)
	// Base64 validates if a string is encoded in Base64
	Base64 = validation.NewStringRuleWithError(RuleTypeBase64, govalidator.IsBase64, ErrBase64)
	// DataURI validates if a string is a valid base64-encoded data URI
	DataURI = validation.NewStringRuleWithError(RuleTypeDataURI, govalidator.IsDataURI, ErrDataURI)
	// E164 validates if a string is a valid ISO3166 Alpha 2 country code
	E164 = validation.NewStringRuleWithError(RuleTypeE164, isE164Number, ErrE164)
	// CountryCode2 validates if a string is a valid ISO3166 Alpha 2 country code
	CountryCode2 = validation.NewStringRuleWithError(RuleTypeCountryCode2, govalidator.IsISO3166Alpha2, ErrCountryCode2)
	// CountryCode3 validates if a string is a valid ISO3166 Alpha 3 country code
	CountryCode3 = validation.NewStringRuleWithError(RuleTypeCountryCode3, govalidator.IsISO3166Alpha3, ErrCountryCode3)
	// CurrencyCode validates if a string is a valid IsISO4217 currency code.
	CurrencyCode = validation.NewStringRuleWithError(RuleTypeCurrencyCode, govalidator.IsISO4217, ErrCurrencyCode)
	// DialString validates if a string is a valid dial string that can be passed to Dial()
	DialString = validation.NewStringRuleWithError(RuleTypeDialString, govalidator.IsDialString, ErrDialString)
	// MAC validates if a string is a MAC address
	MAC = validation.NewStringRuleWithError(RuleTypeMAC, govalidator.IsMAC, ErrMac)
	// IP validates if a string is a valid IP address (either version 4 or 6)
	IP = validation.NewStringRuleWithError(RuleTypeIP, govalidator.IsIP, ErrIP)
	// IPv4 validates if a string is a valid version 4 IP address
	IPv4 = validation.NewStringRuleWithError(RuleTypeIPv4, govalidator.IsIPv4, ErrIPv4)
	// IPv6 validates if a string is a valid version 6 IP address
	IPv6 = validation.NewStringRuleWithError(RuleTypeIPv6, govalidator.IsIPv6, ErrIPv6)
	// Subdomain validates if a string is valid subdomain
	Subdomain = validation.NewStringRuleWithError(RuleTypeSubdomain, isSubdomain, ErrSubdomain)
	// Domain validates if a string is valid domain
	Domain = validation.NewStringRuleWithError(RuleTypeDomain, isDomain, ErrDomain)
	// DNSName validates if a string is valid DNS name
	DNSName = validation.NewStringRuleWithError(RuleTypeDNSName, govalidator.IsDNSName, ErrDNSName)
	// Host validates if a string is a valid IP (both v4 and v6) or a valid DNS name
	Host = validation.NewStringRuleWithError(RuleTypeHost, govalidator.IsHost, ErrHost)
	// Port validates if a string is a valid port number
	Port = validation.NewStringRuleWithError(RuleTypePort, govalidator.IsPort, ErrPort)
	// MongoID validates if a string is a valid Mongo ID
	MongoID = validation.NewStringRuleWithError(RuleTypeMongoID, govalidator.IsMongoID, ErrMongoID)
	// Latitude validates if a string is a valid latitude
	Latitude = validation.NewStringRuleWithError(RuleTypeLatitude, govalidator.IsLatitude, ErrLatitude)
	// Longitude validates if a string is a valid longitude
	Longitude = validation.NewStringRuleWithError(RuleTypeLongitude, govalidator.IsLongitude, ErrLongitude)
	// SSN validates if a string is a social security number (SSN)
	SSN = validation.NewStringRuleWithError(RuleTypeSSN, govalidator.IsSSN, ErrSSN)
	// Semver validates if a string is a valid semantic version
	Semver = validation.NewStringRuleWithError(RuleTypeSemver, govalidator.IsSemver, ErrSemver)
)

var (
	reDigit = regexp.MustCompile("^[0-9]+$")
	// Subdomain regex source: https://stackoverflow.com/a/7933253
	reSubdomain = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?$`)
	// E164 regex source: https://stackoverflow.com/a/23299989
	reE164 = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	// Domain regex source: https://stackoverflow.com/a/7933253
	// Slightly modified: Removed 255 max length validation since Go regex does not
	// support lookarounds. More info: https://stackoverflow.com/a/38935027
	reDomain = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-zA-Z]{1,63}| xn--[a-z0-9]{1,59})$`)
)

func isISBN(value string) bool {
	return govalidator.IsISBN(value, 10) || govalidator.IsISBN(value, 13)
}

func isDigit(value string) bool {
	return reDigit.MatchString(value)
}

func isE164Number(value string) bool {
	return reE164.MatchString(value)
}

func isSubdomain(value string) bool {
	return reSubdomain.MatchString(value)
}

func isDomain(value string) bool {
	if len(value) > 255 {
		return false
	}

	return reDomain.MatchString(value)
}

func isUTFNumeric(value string) bool {
	for _, c := range value {
		if unicode.IsNumber(c) == false {
			return false
		}
	}
	return true
}

func init() {
	validation.RegisterRule(Email)
	validation.RegisterRule(EmailFormat)
	validation.RegisterRule(URL)
	validation.RegisterRule(RequestURL)
	validation.RegisterRule(RequestURI)
	validation.RegisterRule(Alpha)
	validation.RegisterRule(Digit)
	validation.RegisterRule(Alphanumeric)
	validation.RegisterRule(UTFLetter)
	validation.RegisterRule(UTFDigit)
	validation.RegisterRule(UTFLetterNumeric)
	validation.RegisterRule(UTFNumeric)
	validation.RegisterRule(LowerCase)
	validation.RegisterRule(UpperCase)
	validation.RegisterRule(Hexadecimal)
	validation.RegisterRule(HexColor)
	validation.RegisterRule(RGBColor)
	validation.RegisterRule(Int)
	validation.RegisterRule(Float)
	validation.RegisterRule(UUIDv3)
	validation.RegisterRule(UUIDv4)
	validation.RegisterRule(UUIDv5)
	validation.RegisterRule(UUID)
	validation.RegisterRule(CreditCard)
	validation.RegisterRule(ISBN10)
	validation.RegisterRule(ISBN13)
	validation.RegisterRule(ISBN)
	validation.RegisterRule(JSON)
	validation.RegisterRule(ASCII)
	validation.RegisterRule(PrintableASCII)
	validation.RegisterRule(Multibyte)
	validation.RegisterRule(FullWidth)
	validation.RegisterRule(HalfWidth)
	validation.RegisterRule(VariableWidth)
	validation.RegisterRule(Base64)
	validation.RegisterRule(DataURI)
	validation.RegisterRule(E164)
	validation.RegisterRule(CountryCode2)
	validation.RegisterRule(CountryCode3)
	validation.RegisterRule(CurrencyCode)
	validation.RegisterRule(DialString)
	validation.RegisterRule(MAC)
	validation.RegisterRule(IP)
	validation.RegisterRule(IPv4)
	validation.RegisterRule(IPv6)
	validation.RegisterRule(Subdomain)
	validation.RegisterRule(Domain)
	validation.RegisterRule(DNSName)
	validation.RegisterRule(Host)
	validation.RegisterRule(Port)
	validation.RegisterRule(MongoID)
	validation.RegisterRule(Latitude)
	validation.RegisterRule(Longitude)
	validation.RegisterRule(SSN)
	validation.RegisterRule(Semver)
}
