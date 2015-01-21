package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

func bootstrapprogressbar_min_js() ([]byte, error) {
	return bindata_read([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0xb4, 0x56,
		0xc1, 0x72, 0xe3, 0x36, 0x0f, 0xbe, 0xff, 0x4f, 0xa1, 0x70, 0xfe, 0x7a,
		0xc8, 0x84, 0xa6, 0xbd, 0xdb, 0xed, 0x4c, 0x47, 0x2e, 0x37, 0x87, 0x1c,
		0x3a, 0x39, 0xec, 0xce, 0x74, 0x9a, 0xbb, 0x87, 0x91, 0x20, 0x89, 0x5b,
		0x99, 0xd4, 0x90, 0x94, 0x9d, 0xd4, 0xd6, 0xb3, 0xf4, 0x61, 0xfa, 0x62,
		0x05, 0x65, 0xd9, 0x96, 0xed, 0xcd, 0x76, 0x2f, 0x3d, 0xc4, 0x22, 0x89,
		0x0f, 0x20, 0x00, 0x02, 0x1f, 0x32, 0xbb, 0xbd, 0x49, 0x9e, 0xad, 0x0d,
		0x3e, 0x38, 0xd5, 0x4c, 0x1b, 0x67, 0x4b, 0x07, 0xde, 0x3f, 0x2b, 0x97,
		0xac, 0xe7, 0xe2, 0x67, 0xf1, 0x21, 0xd9, 0x25, 0x0f, 0xb6, 0x79, 0x75,
		0xba, 0xac, 0x42, 0x42, 0x33, 0x96, 0xbc, 0x9f, 0xbf, 0x7b, 0x3f, 0xc5,
		0x9f, 0x0f, 0xc9, 0xef, 0x01, 0x9a, 0x4a, 0x99, 0xe4, 0x57, 0x67, 0xff,
		0xfe, 0x0b, 0x71, 0x9f, 0x1e, 0x9f, 0x92, 0x5a, 0x67, 0x60, 0x3c, 0xe0,
		0xae, 0x0a, 0xa1, 0x49, 0x67, 0xb3, 0xcd, 0x66, 0x23, 0x56, 0xda, 0xe4,
		0x79, 0xeb, 0x83, 0xc8, 0xec, 0x2a, 0xb9, 0x9d, 0xfd, 0xef, 0xa6, 0x68,
		0x4d, 0x16, 0xb4, 0x35, 0x34, 0xb0, 0x2d, 0x69, 0x11, 0x8e, 0xb7, 0xeb,
		0x2c, 0x90, 0xc5, 0x1a, 0xef, 0x05, 0x79, 0x14, 0x1b, 0xae, 0xd8, 0x36,
		0x54, 0xda, 0x8b, 0xff, 0x43, 0x0d, 0x2b, 0x30, 0x41, 0x06, 0x6a, 0x18,
		0xef, 0x8f, 0x6c, 0x13, 0x31, 0x5e, 0x06, 0x01, 0x2f, 0x01, 0x4c, 0x4e,
		0xb7, 0x1d, 0x07, 0x91, 0x43, 0xa1, 0xda, 0x3a, 0x78, 0xd4, 0xec, 0x16,
		0xa7, 0xad, 0xdc, 0x62, 0x7c, 0xc6, 0xeb, 0xa8, 0xb2, 0xcc, 0xa1, 0x56,
		0xaf, 0xe9, 0x8f, 0xf3, 0x39, 0x77, 0x50, 0x60, 0xb8, 0xd5, 0xd2, 0x37,
		0x00, 0x79, 0xfa, 0xd3, 0x9c, 0xe7, 0xda, 0x37, 0x28, 0x5c, 0x06, 0xb4,
		0x99, 0x12, 0x63, 0x0d, 0x10, 0x8e, 0x0e, 0x2e, 0x1b, 0x70, 0x18, 0x58,
		0x50, 0x25, 0xa4, 0x37, 0x73, 0x3e, 0xec, 0x96, 0x85, 0x75, 0x2b, 0x15,
		0xd2, 0x71, 0x38, 0x0e, 0x42, 0xeb, 0x4c, 0x12, 0xee, 0xc8, 0x0f, 0xa4,
		0xe3, 0x6a, 0x65, 0xdb, 0xaf, 0xe1, 0x38, 0x8c, 0x91, 0xc9, 0x2c, 0x21,
		0x77, 0xd0, 0xf1, 0xb6, 0xc9, 0x55, 0x80, 0x34, 0x08, 0x63, 0x6d, 0xc3,
		0x73, 0xbc, 0xfc, 0xb0, 0x2e, 0x94, 0xae, 0x87, 0x75, 0x0c, 0x12, 0x9f,
		0x29, 0xd8, 0xf0, 0xda, 0x80, 0x38, 0x45, 0x75, 0xca, 0x1a, 0xdb, 0xc6,
		0x34, 0x1a, 0x79, 0x96, 0x38, 0xae, 0xa4, 0x11, 0x8d, 0x72, 0xb8, 0xa4,
		0x8c, 0xfb, 0x41, 0xf8, 0xac, 0xb2, 0x3f, 0xfa, 0x58, 0xb9, 0x1b, 0x4e,
		0x0a, 0x67, 0xd1, 0xe1, 0xfe, 0x48, 0xcb, 0x71, 0x9e, 0xb9, 0x95, 0xa8,
		0xee, 0xe1, 0x11, 0x0d, 0x18, 0xa1, 0x42, 0x70, 0x94, 0xa0, 0xbb, 0x6a,
		0x7a, 0x72, 0xa1, 0xb4, 0xaa, 0x26, 0x8c, 0xf1, 0xea, 0x1a, 0xa9, 0x9c,
		0x56, 0xd3, 0xb5, 0xaa, 0x5b, 0xc0, 0x62, 0x40, 0xcc, 0x6e, 0x87, 0xb9,
		0xfe, 0x26, 0x4c, 0xbd, 0xf4, 0xb0, 0x77, 0xf8, 0x4a, 0x85, 0x54, 0xa2,
		0x52, 0xfe, 0xa1, 0x56, 0xde, 0x53, 0xb2, 0x06, 0x17, 0x74, 0x16, 0x6f,
		0xe2, 0x8d, 0xd4, 0x62, 0x9f, 0xb4, 0xc9, 0x84, 0x1c, 0xe2, 0x27, 0x52,
		0xc6, 0xd4, 0xd8, 0x22, 0x39, 0x08, 0xef, 0x0f, 0x8b, 0xf4, 0x54, 0x10,
		0xc3, 0x09, 0x6f, 0xd1, 0x44, 0xcc, 0xf4, 0x1b, 0x06, 0xa2, 0xe8, 0x7e,
		0xff, 0x19, 0x2b, 0xc7, 0x3d, 0xcf, 0x50, 0x35, 0x3e, 0xcc, 0x1b, 0xaa,
		0x51, 0x74, 0xbf, 0xff, 0x8c, 0x55, 0xe3, 0x7e, 0xa1, 0x0b, 0xaa, 0xfd,
		0x67, 0xf5, 0x99, 0x5a, 0xc6, 0x86, 0x3a, 0x58, 0x5b, 0x9d, 0x27, 0xd9,
		0x57, 0x73, 0x9a, 0x18, 0x1b, 0x12, 0x0f, 0x81, 0xb0, 0xbe, 0x41, 0x6a,
		0xf9, 0x49, 0x85, 0x4a, 0x38, 0x2c, 0xad, 0x9c, 0x62, 0x7e, 0x6e, 0xa9,
		0x9d, 0x56, 0x6c, 0x46, 0x73, 0xfc, 0x65, 0xd1, 0x34, 0x89, 0xc5, 0x09,
		0x0e, 0x9d, 0x89, 0xc1, 0x8d, 0x2a, 0x7a, 0x32, 0xb9, 0xf1, 0xf8, 0xe7,
		0x0e, 0x4d, 0x75, 0x7c, 0x7e, 0x89, 0x05, 0x41, 0xc9, 0x2f, 0xbe, 0x51,
		0xe6, 0x23, 0x61, 0x42, 0xe5, 0xf9, 0x90, 0xec, 0x11, 0x21, 0x4c, 0x23,
		0x7a, 0x1a, 0xd1, 0x88, 0x68, 0x1c, 0x34, 0xd8, 0x73, 0x4f, 0x96, 0xaa,
		0xa1, 0x1d, 0x47, 0x95, 0x23, 0xdd, 0x77, 0x58, 0xeb, 0xe1, 0xd7, 0xe6,
		0xcc, 0x3e, 0xc6, 0x72, 0x51, 0xdc, 0xd3, 0x12, 0xdf, 0x3d, 0x8b, 0x7a,
		0x15, 0x44, 0xfa, 0xc1, 0x07, 0xf7, 0xfd, 0x7e, 0xbb, 0xdf, 0xa7, 0x25,
		0x27, 0xb5, 0x36, 0x30, 0x1d, 0xc4, 0x69, 0xd9, 0x31, 0xee, 0xfe, 0x15,
		0x11, 0xe8, 0x06, 0xf9, 0xc8, 0x6e, 0x98, 0x40, 0x57, 0xf4, 0x9f, 0x40,
		0x47, 0x8d, 0xf3, 0x9f, 0xdc, 0xd8, 0x31, 0x96, 0x9e, 0x62, 0xd9, 0xe8,
		0x3c, 0x54, 0xe4, 0xa8, 0xd6, 0x6f, 0xbf, 0xdf, 0xaf, 0xb7, 0xb4, 0xf1,
		0x8e, 0x0e, 0x2b, 0xe4, 0x49, 0xaf, 0xc0, 0xb6, 0x81, 0x5e, 0x70, 0x01,
		0xf2, 0x0d, 0xcf, 0x78, 0xc9, 0x97, 0x98, 0x55, 0x73, 0x16, 0x20, 0xaf,
		0x23, 0x51, 0xb1, 0xd4, 0x8c, 0xcd, 0x0f, 0x87, 0xfd, 0x43, 0xbc, 0x48,
		0x34, 0xfb, 0x18, 0xcb, 0x09, 0xdb, 0x72, 0x6c, 0x17, 0xdf, 0x27, 0x43,
		0x4a, 0xd9, 0xdb, 0x41, 0x4a, 0x89, 0x2e, 0x1e, 0x36, 0x18, 0x6f, 0x94,
		0xf5, 0xd6, 0x06, 0xd1, 0xb0, 0xc6, 0x30, 0x2f, 0xab, 0x37, 0x9b, 0x95,
		0x8c, 0xc3, 0xf8, 0xb4, 0xba, 0xc3, 0xb3, 0xdb, 0x7d, 0x3d, 0xf3, 0xf0,
		0x51, 0x62, 0x87, 0xd1, 0x20, 0x6b, 0x04, 0x59, 0xde, 0x46, 0xfe, 0xcf,
		0x6a, 0x50, 0xee, 0xe8, 0xd4, 0x0b, 0xa2, 0xf6, 0x54, 0x7d, 0x73, 0x5d,
		0xf0, 0x74, 0x19, 0x39, 0xe2, 0x8c, 0xc1, 0xb1, 0x29, 0xcf, 0x09, 0x1c,
		0x79, 0x3b, 0xd5, 0xe2, 0x8c, 0xab, 0x29, 0xf0, 0x9c, 0x57, 0x68, 0xb7,
		0xd0, 0x75, 0x7d, 0xdd, 0x48, 0x98, 0xc5, 0xf8, 0xa1, 0x4b, 0x96, 0x7e,
		0xa3, 0xd9, 0xa8, 0x3f, 0xa0, 0xf0, 0xbd, 0x86, 0x15, 0xfa, 0x7a, 0x4d,
		0x76, 0xc6, 0x6e, 0x08, 0x8e, 0x04, 0xde, 0xe0, 0x64, 0x30, 0xac, 0xe3,
		0x5a, 0x9c, 0x0d, 0xa6, 0xfe, 0xe4, 0x72, 0x7e, 0xe1, 0x78, 0x1b, 0x68,
		0x5e, 0x14, 0x46, 0x8c, 0x3a, 0x6b, 0x71, 0x79, 0x30, 0x9a, 0xa7, 0xa7,
		0xa9, 0x13, 0x5b, 0x16, 0x54, 0x56, 0x5d, 0x96, 0x8a, 0xc2, 0xd6, 0x8d,
		0xc2, 0x38, 0x23, 0x94, 0x88, 0x64, 0x44, 0xc9, 0xb3, 0x1f, 0x9b, 0x8b,
		0xd5, 0x27, 0x89, 0x7d, 0xfe, 0x02, 0x38, 0xb1, 0x8f, 0x7c, 0x67, 0x26,
		0x13, 0xb3, 0xf0, 0xbb, 0xdd, 0x1b, 0x3a, 0x68, 0xcd, 0xc0, 0x26, 0x81,
		0xde, 0x36, 0x77, 0x2c, 0x36, 0xd6, 0x29, 0x22, 0x8a, 0xf5, 0xdb, 0xf1,
		0x4b, 0xbf, 0xc5, 0x03, 0x8e, 0x9d, 0xe0, 0xda, 0x2c, 0x58, 0x27, 0xe1,
		0x5a, 0x6c, 0x2c, 0x02, 0x0a, 0xfc, 0x8f, 0x23, 0x8c, 0x67, 0xdf, 0x21,
		0xc0, 0xcb, 0x24, 0x98, 0x9e, 0xa6, 0xba, 0x6e, 0xe8, 0x32, 0xf1, 0xe5,
		0xb7, 0x16, 0xdc, 0x2b, 0x5b, 0xfc, 0x13, 0x00, 0x00, 0xff, 0xff, 0x88,
		0xdf, 0x7e, 0xc4, 0x01, 0x09, 0x00, 0x00,
	},
		"bootstrapProgressbar.min.js",
	)
}

func list_tmpl() ([]byte, error) {
	return bindata_read([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0x9c, 0x54,
		0x4d, 0x6f, 0x13, 0x31, 0x10, 0x3d, 0x97, 0x5f, 0xe1, 0xfa, 0x92, 0x53,
		0xd7, 0x0a, 0xbd, 0xa0, 0xca, 0xbb, 0x12, 0x50, 0x90, 0x90, 0x10, 0x54,
		0x2a, 0x1c, 0xe0, 0x82, 0xbc, 0xeb, 0x49, 0xd7, 0xc1, 0xbb, 0x36, 0xf6,
		0x6c, 0x68, 0x88, 0xf2, 0xdf, 0xf1, 0xc7, 0x7e, 0xd0, 0xd0, 0x43, 0xd5,
		0x4b, 0xfc, 0x66, 0xc6, 0xf3, 0xe6, 0xed, 0xb3, 0x1d, 0x7e, 0x7e, 0xfd,
		0xf9, 0xed, 0x97, 0x6f, 0x37, 0xef, 0x48, 0x8b, 0x9d, 0xae, 0x5e, 0xf0,
		0xbc, 0x9c, 0xf1, 0x16, 0x84, 0x0c, 0xeb, 0x19, 0xef, 0x00, 0x05, 0x69,
		0x5a, 0xe1, 0x3c, 0x60, 0x49, 0x07, 0xdc, 0x5c, 0xbc, 0xa2, 0xa9, 0x80,
		0x0a, 0x35, 0x54, 0x1b, 0xa5, 0x41, 0x3a, 0x63, 0x39, 0xcb, 0x71, 0xac,
		0x68, 0xd5, 0xff, 0x24, 0x0e, 0x74, 0x49, 0x3d, 0xee, 0x35, 0xf8, 0x16,
		0x00, 0x29, 0x69, 0x1d, 0x6c, 0x4a, 0xda, 0x22, 0x5a, 0x7f, 0xc5, 0x58,
		0x27, 0xee, 0x1b, 0xd9, 0x17, 0xb5, 0x31, 0xe8, 0xd1, 0x09, 0x1b, 0x83,
		0xc6, 0x74, 0x6c, 0x4e, 0xb0, 0xcb, 0xe2, 0xb2, 0x58, 0xb3, 0xc6, 0xfb,
		0x25, 0x57, 0x74, 0x2a, 0xec, 0xf2, 0x3e, 0x0b, 0xf0, 0x8d, 0x53, 0x16,
		0x89, 0x77, 0x4d, 0x49, 0x19, 0x0b, 0x04, 0x5b, 0x5f, 0x34, 0xda, 0x0c,
		0x72, 0xa3, 0x85, 0x83, 0xc4, 0x26, 0xb6, 0xe2, 0x9e, 0x69, 0x55, 0x7b,
		0xb6, 0xfd, 0x35, 0x80, 0xdb, 0xb3, 0x97, 0xc5, 0x3a, 0x90, 0xe6, 0x20,
		0xb1, 0x6d, 0x3d, 0x25, 0xb8, 0xb7, 0x50, 0x52, 0x84, 0x7b, 0x64, 0x5b,
		0xb1, 0x13, 0x99, 0x97, 0x56, 0x9c, 0x65, 0xf4, 0xc8, 0xb0, 0xa7, 0xaa,
		0xdf, 0x9e, 0x8a, 0x7f, 0xce, 0xb8, 0x99, 0xe1, 0xc6, 0x99, 0x3b, 0x07,
		0xde, 0xd7, 0xc2, 0x3d, 0x9b, 0x6c, 0xb0, 0xda, 0x08, 0x39, 0xa8, 0x27,
		0x36, 0x73, 0x36, 0x5e, 0x04, 0x5e, 0x1b, 0xb9, 0x4f, 0x6c, 0x52, 0xed,
		0x48, 0xa3, 0x85, 0xf7, 0x25, 0x6d, 0x4c, 0x8f, 0x42, 0xf5, 0xe0, 0xf2,
		0x81, 0xb4, 0xeb, 0xea, 0x6b, 0xa2, 0x27, 0xef, 0xc3, 0xad, 0x08, 0xad,
		0xeb, 0x94, 0xb6, 0xf1, 0xf7, 0xec, 0x16, 0x34, 0x34, 0x98, 0x2a, 0x57,
		0x31, 0xe6, 0xaa, 0xb7, 0x03, 0x66, 0x09, 0xab, 0x78, 0x8b, 0x56, 0x44,
		0xc9, 0x72, 0xf5, 0x23, 0xc1, 0xea, 0xbf, 0x1d, 0xf5, 0x80, 0x68, 0xfa,
		0x71, 0x8f, 0x1f, 0xea, 0x4e, 0xe1, 0x8a, 0xec, 0x84, 0x1e, 0x42, 0x2d,
		0x0f, 0x3d, 0x4f, 0x5d, 0x9c, 0xd9, 0x53, 0x95, 0x76, 0xb4, 0x8d, 0x4c,
		0xe0, 0x22, 0xb8, 0xa9, 0x2c, 0x48, 0x22, 0x1a, 0x54, 0x3b, 0xa0, 0x79,
		0xda, 0x23, 0x1d, 0x17, 0xc1, 0x69, 0xf2, 0x6f, 0x30, 0x75, 0x52, 0xe2,
		0x8c, 0x86, 0x65, 0x63, 0x28, 0x45, 0xdb, 0x02, 0x45, 0xd6, 0x90, 0xc0,
		0x68, 0xc9, 0x47, 0xe5, 0x91, 0x98, 0x0d, 0x89, 0x1f, 0xe6, 0x67, 0x53,
		0x50, 0xd4, 0x1a, 0xa6, 0x79, 0x29, 0x18, 0x5f, 0xd5, 0xf4, 0xf0, 0x02,
		0x74, 0x69, 0x0d, 0x40, 0x56, 0x9f, 0x44, 0x17, 0x0c, 0x45, 0xb9, 0x64,
		0x6e, 0xd5, 0x9f, 0x25, 0x13, 0x80, 0xcb, 0x93, 0xe7, 0x7e, 0x8e, 0xd3,
		0x89, 0x1d, 0x0e, 0x4e, 0xf4, 0x77, 0x40, 0x8a, 0xe3, 0x31, 0xee, 0x3d,
		0x1c, 0xd4, 0x86, 0x14, 0x1f, 0xfc, 0xb5, 0x72, 0x53, 0x02, 0xb4, 0x87,
		0x8c, 0x1f, 0x0c, 0xe5, 0x62, 0x7c, 0xb4, 0x49, 0x3a, 0x3b, 0x1c, 0x8a,
		0x28, 0xe3, 0x78, 0xa4, 0xd5, 0x0c, 0x39, 0x13, 0xd5, 0x43, 0x5d, 0xa1,
		0x14, 0xa5, 0x1d, 0x8f, 0xe4, 0xcd, 0x1e, 0xe3, 0x07, 0x9f, 0x48, 0x8c,
		0xe3, 0x7a, 0x99, 0xa6, 0x2d, 0x28, 0x14, 0xe7, 0xfb, 0xc5, 0x92, 0x1b,
		0xb3, 0x7d, 0xdf, 0x95, 0x8d, 0xee, 0x09, 0xad, 0x4f, 0x1c, 0x9c, 0xc4,
		0x31, 0x69, 0x7e, 0xf7, 0xf1, 0x02, 0xbc, 0xd6, 0x9a, 0x56, 0xd7, 0x63,
		0x10, 0x85, 0x2d, 0x47, 0xc1, 0x59, 0xa6, 0x0f, 0xcd, 0xf1, 0xff, 0xed,
		0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x22, 0x91, 0xbb, 0x6d, 0xf6, 0x04,
		0x00, 0x00,
	},
		"list.tmpl",
	)
}

func uploadui_js() ([]byte, error) {
	return bindata_read([]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x00, 0xff, 0x7c, 0x54,
		0x51, 0x6f, 0xd3, 0x30, 0x10, 0x7e, 0xe7, 0x57, 0x9c, 0xa6, 0x49, 0x76,
		0xa0, 0x33, 0x45, 0xe2, 0x69, 0x53, 0x5f, 0x60, 0x43, 0x80, 0x36, 0x8a,
		0xe8, 0x90, 0x90, 0xa6, 0xa9, 0x72, 0x93, 0x6b, 0x1b, 0xcd, 0xb5, 0x83,
		0x7d, 0xd9, 0xa8, 0xaa, 0xfe, 0x77, 0xce, 0x49, 0x9a, 0x46, 0x69, 0xc0,
		0x0f, 0x95, 0x7b, 0xf9, 0xee, 0xbb, 0xef, 0x2e, 0xf7, 0x25, 0x73, 0x69,
		0xb9, 0x41, 0x4b, 0x4a, 0x67, 0xd9, 0xcd, 0x33, 0x5f, 0x6e, 0xf3, 0x40,
		0x68, 0xd1, 0xcb, 0xb3, 0xeb, 0xe9, 0xdd, 0x47, 0x67, 0x29, 0xc6, 0x9c,
		0xce, 0x30, 0x3b, 0x1b, 0xc1, 0xb2, 0xb4, 0x29, 0xe5, 0xce, 0x4a, 0x8c,
		0xd0, 0x04, 0x76, 0xf0, 0x0a, 0xf8, 0x3c, 0x6b, 0x0f, 0xf3, 0x50, 0x2e,
		0x36, 0x39, 0xc1, 0x04, 0x5a, 0xca, 0x15, 0xd2, 0x8d, 0xc1, 0x78, 0xfd,
		0xb0, 0xfd, 0x92, 0x49, 0xd1, 0x40, 0x44, 0x32, 0xaa, 0xd3, 0xe6, 0xcb,
		0xdc, 0xe0, 0x7f, 0x13, 0x22, 0x40, 0x24, 0x57, 0x15, 0xba, 0xad, 0x74,
		0x5e, 0x2c, 0x38, 0xe9, 0x5c, 0x0a, 0x55, 0x78, 0xb7, 0xf2, 0x18, 0x02,
		0xb4, 0xb7, 0x8b, 0x85, 0xf6, 0x31, 0xa1, 0x05, 0x97, 0x85, 0x61, 0xf1,
		0x8c, 0x6f, 0xa5, 0x27, 0xbb, 0xfa, 0x69, 0x3c, 0xf9, 0x52, 0x56, 0x35,
		0x54, 0xfc, 0x09, 0xca, 0xa0, 0x5d, 0xd1, 0x1a, 0x26, 0x93, 0x09, 0x8c,
		0x19, 0x06, 0x9d, 0xe3, 0x91, 0x4a, 0x6f, 0xaf, 0xda, 0xd8, 0xfe, 0x48,
		0x12, 0xcb, 0x64, 0x9a, 0x34, 0x17, 0xb1, 0xf8, 0x02, 0x9f, 0x9c, 0xdf,
		0x5c, 0xf3, 0x5f, 0x99, 0x1c, 0xd1, 0xf1, 0xb1, 0xd2, 0x45, 0x81, 0x96,
		0xbb, 0x5a, 0xd6, 0x9a, 0xc4, 0x08, 0x3a, 0xb5, 0x1f, 0xc6, 0x8f, 0x07,
		0xd9, 0x07, 0x4e, 0x8f, 0xbf, 0x4b, 0x0c, 0xd4, 0xd0, 0xfe, 0xba, 0xbb,
		0xfd, 0x4c, 0x54, 0xfc, 0xa8, 0x83, 0x5d, 0xf2, 0x06, 0xa7, 0x9c, 0xf5,
		0xa8, 0xb3, 0x6d, 0x20, 0x4d, 0x98, 0xae, 0xb5, 0x5d, 0x61, 0xaf, 0xed,
		0x6e, 0x3f, 0xdc, 0xf9, 0x21, 0xaf, 0xca, 0x9a, 0xc5, 0x2c, 0xee, 0x1c,
		0xde, 0xf7, 0x80, 0xf1, 0x90, 0xdf, 0xc2, 0x69, 0xf4, 0xa8, 0x33, 0x14,
		0x5c, 0xe9, 0xeb, 0x6c, 0xfa, 0x4d, 0x15, 0xda, 0x07, 0xec, 0x30, 0x87,
		0xc2, 0xd9, 0x80, 0x1d, 0xb1, 0xed, 0xfc, 0x20, 0xd5, 0x94, 0xae, 0x41,
		0xe2, 0x40, 0xbd, 0x1e, 0xf3, 0x30, 0x20, 0x9e, 0xd8, 0x6b, 0x19, 0x2e,
		0x41, 0xa0, 0xf7, 0xce, 0x8b, 0xd1, 0x3f, 0x81, 0xf1, 0x05, 0x30, 0xec,
		0xa7, 0x7d, 0xb2, 0xee, 0xc5, 0x42, 0x05, 0x07, 0x97, 0xa6, 0xa5, 0xf7,
		0x98, 0x5d, 0xc2, 0x83, 0x80, 0x37, 0xd0, 0x57, 0x7d, 0x8f, 0x7f, 0x88,
		0xc3, 0xe2, 0x51, 0x0c, 0xd2, 0xee, 0x07, 0x9a, 0x3a, 0x89, 0x68, 0x83,
		0x9e, 0x64, 0x64, 0x54, 0xb5, 0xd6, 0x48, 0xc8, 0x42, 0xaa, 0x72, 0x1c,
		0x8c, 0xba, 0x7a, 0xd3, 0x39, 0x92, 0xec, 0x3b, 0x0b, 0x71, 0x10, 0x57,
		0xef, 0xce, 0xa9, 0x5b, 0xc5, 0xc1, 0x01, 0xa2, 0xeb, 0xd2, 0xde, 0x6c,
		0xd9, 0x39, 0x4a, 0x13, 0x31, 0x3a, 0xd6, 0xbd, 0x20, 0xaf, 0x6d, 0xc8,
		0x23, 0x72, 0xe5, 0xb4, 0xe1, 0xc4, 0x3b, 0x4d, 0x6b, 0x95, 0x62, 0x6e,
		0x24, 0x2a, 0x53, 0x79, 0xfe, 0x35, 0xbc, 0x1b, 0x8f, 0xdf, 0xa2, 0x22,
		0x47, 0xda, 0x24, 0x49, 0xeb, 0x33, 0xb6, 0x99, 0xdc, 0x65, 0x79, 0x28,
		0x8c, 0xde, 0xce, 0x89, 0x07, 0xc5, 0x3d, 0xf1, 0x1e, 0x1b, 0xb1, 0xef,
		0x74, 0xb3, 0x67, 0x29, 0xda, 0x54, 0xaf, 0xff, 0x74, 0x59, 0xd9, 0x0b,
		0x52, 0x7c, 0x9f, 0xce, 0xee, 0xb9, 0xae, 0x68, 0x1c, 0x31, 0xb0, 0xd4,
		0x21, 0x5a, 0xa6, 0x33, 0xa5, 0xc6, 0x77, 0xcd, 0xa7, 0x64, 0x60, 0x0e,
		0xa9, 0xc9, 0xd3, 0x27, 0xe6, 0xac, 0x29, 0x39, 0x8b, 0x15, 0xfd, 0x0d,
		0x00, 0x00, 0xff, 0xff, 0xc6, 0xbd, 0xd8, 0xc5, 0xe5, 0x04, 0x00, 0x00,
	},
		"uploadui.js",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"bootstrapProgressbar.min.js": bootstrapprogressbar_min_js,
	"list.tmpl": list_tmpl,
	"uploadui.js": uploadui_js,
}
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"bootstrapProgressbar.min.js": &_bintree_t{bootstrapprogressbar_min_js, map[string]*_bintree_t{
	}},
	"list.tmpl": &_bintree_t{list_tmpl, map[string]*_bintree_t{
	}},
	"uploadui.js": &_bintree_t{uploadui_js, map[string]*_bintree_t{
	}},
}}
