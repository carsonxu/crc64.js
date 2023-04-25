// Copyright 2015, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

// Package hashmerge provides functionality for merging hashes.
package main

import (
	"fmt"
	"hash/crc64"
)

// The origin of the CombineAdler32, CombineCRC32, and CombineCRC64 functions
// in this package is the adler32_combine, crc32_combine, gf2_matrix_times,
// and gf2_matrix_square functions found in the zlib library and was translated
// from C to Go. Thanks goes to the authors of zlib:
//	Mark Adler and Jean-loup Gailly.
//
// See the following:
//	http://www.zlib.net/
//	https://github.com/madler/zlib/blob/master/adler32.c
//	https://github.com/madler/zlib/blob/master/crc32.c
//	https://stackoverflow.com/questions/23122312/crc-calculation-of-a-mostly-static-data-stream/23126768#23126768
//
// ====================================================
// Copyright (C) 1995-2013 Jean-loup Gailly and Mark Adler
//
// This software is provided 'as-is', without any express or implied
// warranty.  In no event will the authors be held liable for any damages
// arising from the use of this software.
//
// Permission is granted to anyone to use this software for any purpose,
// including commercial applications, and to alter it and redistribute it
// freely, subject to the following restrictions:
//
// 1. The origin of this software must not be misrepresented; you must not
//    claim that you wrote the original software. If you use this software
//    in a product, an acknowledgment in the product documentation would be
//    appreciated but is not required.
// 2. Altered source versions must be plainly marked as such, and must not be
//    misrepresented as being the original software.
// 3. This notice may not be removed or altered from any source distribution.
//
// Jean-loup Gailly        Mark Adler
// jloup@gzip.org          madler@alumni.caltech.edu
// ====================================================

// CombineAdler32 combines two Adler-32 checksums together.
// Let AB be the string concatenation of two strings A and B. Then Combine
// computes the checksum of AB given only the checksum of A, the checksum of B,
// and the length of B:
//
//	adler32.Checksum(AB) == CombineAdler32(adler32.Checksum(A), adler32.Checksum(B), len(B))
func CombineAdler32(adler1, adler2 uint32, len2 int64) uint32 {
	if len2 < 0 {
		panic("hashmerge: negative length")
	}

	const mod = 65521
	var sum1, sum2, rem uint32
	rem = uint32(len2 % mod)
	sum1 = adler1 & 0xffff
	sum2 = rem * sum1
	sum2 %= mod
	sum1 += (adler2 & 0xffff) + mod - 1
	sum2 += (adler1 >> 16) + (adler2 >> 16) + mod - rem
	if sum1 >= mod {
		sum1 -= mod
	}
	if sum1 >= mod {
		sum1 -= mod
	}
	if sum2 >= mod<<1 {
		sum2 -= mod << 1
	}
	if sum2 >= mod {
		sum2 -= mod
	}
	return sum1 | (sum2 << 16)
}

func CombineCRC64(poly, crc1, crc2 uint64, len2 int64) uint64 {
	if len2 < 0 {
		panic("hashmerge: negative length")
	}

	// Translation of gf2_matrix_times from zlib.
	var matrixMult = func(mat *[64]uint64, vec uint64) uint64 {
		var sum uint64
		for n := 0; n < 64 && vec > 0; n++ {
			if vec&1 > 0 {
				sum ^= mat[n]
			}
			vec >>= 1
		}
		return sum
	}

	// Translation of gf2_matrix_square from zlib.
	var matrixSquare = func(square, mat *[64]uint64) {
		for n := 0; n < 64; n++ {
			square[n] = matrixMult(mat, mat[n])
		}
	}

	// Even and odd power-of-two zeros operators.
	var even, odd [64]uint64

	// Put operator for one zero bit in odd.
	var row uint64 = 1
	odd[0] = poly
	for n := 1; n < 64; n++ {
		odd[n] = row
		row <<= 1
	}

	// Put operator for two zero bits in even.
	matrixSquare(&even, &odd)

	// Put operator for four zero bits in odd.
	matrixSquare(&odd, &even)

	// Apply len2 zeros to crc1 (first square will put the operator for one
	// zero byte, eight zero bits, in even).
	for {
		// Apply zeros operator for this bit of len2.
		matrixSquare(&even, &odd)
		if len2&1 > 0 {
			crc1 = matrixMult(&even, crc1)
		}
		len2 >>= 1
		if len2 == 0 {
			break
		}

		// Another iteration of the loop with odd and even swapped.
		matrixSquare(&odd, &even)
		if len2&1 > 0 {
			crc1 = matrixMult(&odd, crc1)
		}
		len2 >>= 1
		if len2 == 0 {
			break
		}
	}
	return crc1 ^ crc2
}

func TestCombineCRC64() {
	var golden = []struct {
		iso, ecma uint64
		in        string
	}{
		{0x0000000000000000, 0x0000000000000000, ""},
		{0x3420000000000000, 0x330284772e652b05, "a"},
		{0x36c4200000000000, 0xbc6573200e84b046, "ab"},
		{0x3776c42000000000, 0x2cd8094a1a277627, "abc"},
		{0x336776c420000000, 0x3c9d28596e5960ba, "abcd"},
		{0x32d36776c4200000, 0x040bdf58fb0895f2, "abcde"},
		{0x3002d36776c42000, 0xd08e9f8545a700f4, "abcdef"},
		{0x31b002d36776c420, 0xec20a3a8cc710e66, "abcdefg"},
		{0x0e21b002d36776c4, 0x67b4f30a647a0c59, "abcdefgh"},
		{0x8b6e21b002d36776, 0x9966f6c89d56ef8e, "abcdefghi"},
		{0x7f5b6e21b002d367, 0x32093a2ecd5773f4, "abcdefghij"},
		{0x8ec0e7c835bf9cdf, 0x8a0825223ea6d221, "Discard medicine more than two years old."},
		{0xc7db1759e2be5ab4, 0x8562c0ac2ab9a00d, "He who has a shady past knows that nice guys finish last."},
		{0xfbf9d9603a6fa020, 0x3ee2a39c083f38b4, "I wouldn't marry him with a ten foot pole."},
		{0xeafc4211a6daa0ef, 0x1f603830353e518a, "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
		{0x3e05b21c7a4dc4da, 0x02fd681d7b2421fd, "The days of the digital watch are numbered.  -Tom Stoppard"},
		{0x5255866ad6ef28a6, 0x790ef2b16a745a41, "Nepal premier won't resign."},
		{0x8a79895be1e9c361, 0x3ef8f06daccdcddf, "For every action there is an equal and opposite government program."},
		{0x8878963a649d4916, 0x049e41b2660b106d, "His money is twice tainted: 'taint yours and 'taint mine."},
		{0xa7b9d53ea87eb82f, 0x561cc0cfa235ac68, "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
		{0xdb6805c0966a2f9c, 0xd4fe9ef082e69f59, "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
		{0xf3553c65dacdadd2, 0xe3b5e46cd8d63a4d, "size:  a.out:  bad magic"},
		{0x9d5e034087a676b9, 0x865aaf6b94f2a051, "The major problem is with sendmail.  -Mark Horton"},
		{0xa6db2d7f8da96417, 0x7eca10d2f8136eb4, "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
		{0x325e00cd2fe819f9, 0xd7dd118c98e98727, "If the enemy is within range, then so are you."},
		{0x88c6600ce58ae4c6, 0x70fb33c119c29318, "It's well we cannot hear the screams/That we create in others' dreams."},
		{0x28c4a3f3b769e078, 0x57c891e39a97d9b7, "You remind me of a TV show, but that's all right: I watch it anyway."},
		{0xa698a34c9d9f1dca, 0xa1f46ba20ad06eb7, "C is as portable as Stonehedge!!"},
		{0xf6c1e2a8c26c5cfc, 0x7ad25fafa1710407, "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
		{0x0d402559dfe9b70c, 0x73cef1666185c13f, "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
		{0xdb6efff26aa94946, 0xb41858f73c389602, "How can you write a big system without C++?  -Paul Glick"},
	}

	var ChecksumECMA = func(data []byte) uint64 {
		return crc64.Checksum(data, crc64.MakeTable(crc64.ECMA))
	}

	for _, g := range golden {
		var splits = []int{
			0 * (len(g.in) / 1),
			1 * (len(g.in) / 4),
			2 * (len(g.in) / 4),
			3 * (len(g.in) / 4),
			1 * (len(g.in) / 1),
		}

		for _, i := range splits {
			p1, p2 := []byte(g.in[:i]), []byte(g.in[i:])
			in1, in2 := g.in[:i], g.in[i:]
			len2 := int64(len(p2))
			if got := CombineCRC64(crc64.ECMA, ChecksumECMA(p1), ChecksumECMA(p2), len2); got != g.ecma {
				fmt.Println("CombineCRC64(ECMA, ChecksumECMA(%q), ChecksumECMA(%q), %d) = 0x%x, want 0x%x",
					in1, in2, len2, got, g.ecma)
			} else {
				fmt.Println("CombineCRC64(ECMA, ChecksumECMA(%q), ChecksumECMA(%q), %d) = 0x%x, want 0x%x",
					in1, in2, len2, got, g.ecma)
			}
		}
	}
}

func crc64_combine(crc1 uint64, crc2 uint64, len2 int64) uint64 {
	return CombineCRC64(crc64.ECMA, crc1, crc2, len2)
}

func main() { // main函数，是程序执行的入口
	// crc64("123"); // 3468660410647627105  0x30232844071cc561n
	// crc64("456"); // 558165746783082364  0x7bf00ca16cbd77cn
	// crc64("123456"); // 318318745347147982  0x46ae5365dc3c8ce
	hash := crc64_combine(3468660410647627105, 558165746783082364, 3)
	fmt.Println("crc64_concat:", hash, 318318745347147982, hash == 318318745347147982) // 在终端打印 Hello World!
}
