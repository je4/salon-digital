package imagecollage

import (
	"errors"
	"fmt"
	"log"
	"sort"
)

// stb_rect_pack.h - v1.01 - public domain - rectangle packing
// Sean Barrett 2014
//
// Useful for e.g. packing rectangular textures into an atlas.
// Does not do rotation.
//
// Before #including,
//
//    #define STB_RECT_PACK_IMPLEMENTATION
//
// in the file that you want to have the implementation.
//
// Not necessarily the awesomest packing method, but better than
// the totally naive one in stb_truetype (which is primarily what
// this is meant to replace).
//
// Has only had a few tests run, may have issues.
//
// More docs to come.
//
// No memory allocations; uses qsort() and assert() from stdlib.
// Can override those by defining STBRP_SORT and STBRP_ASSERT.
//
// This library currently uses the Skyline Bottom-Left algorithm.
//
// Please note: better rectangle packers are welcome! Please
// implement them to the same API, but with a different init
// function.
//
// Credits
//
//  Library
//    Sean Barrett
//  Minor features
//    Martins Mozeiko
//    github:IntellectualKitty
//
//  Bugfixes / warning fixes
//    Jeremy Jaussaud
//    Fabian Giesen
//
// Version history:
//
//     1.01  (2021-07-11)  always use large rect mode, expose STBRP__MAXVAL in public section
//     1.00  (2019-02-25)  avoid small space waste; gracefully fail too-wide rectangles
//     0.99  (2019-02-07)  warning fixes
//     0.11  (2017-03-03)  return packing success/fail result
//     0.10  (2016-10-25)  remove cast-away-const to avoid warnings
//     0.09  (2016-08-27)  fix compiler warnings
//     0.08  (2015-09-13)  really fix bug with empty rects (w=0 or h=0)
//     0.07  (2015-09-13)  fix bug with empty rects (w=0 or h=0)
//     0.06  (2015-04-15)  added STBRP_SORT to allow replacing qsort
//     0.05:  added STBRP_ASSERT to allow replacing assert
//     0.04:  fixed minor bug in STBRP_LARGE_RECTS support
//     0.01:  initial release
//
// LICENSE
//
//   See end of file for license information.
//
//   https://github.com/nothings/stb/blob/master/stb_rect_pack.h

//////////////////////////////////////////////////////////////////////////////
//
//       INCLUDE SECTION
//

type STBRPCoord int

const STBRP__MAXVAL = 0x7fffffff

// Mostly for internal use, but this is the maximum supported coordinate value.

// STBRP_DEF int STBRPPackRects (stbrp_context *context, STBRPRect *rects, int num_rects);
// Assign packed locations to rectangles. The rectangles are of type
// 'STBRPRect' defined below, stored in the array 'rects', and there
// are 'num_rects' many of them.
//
// Rectangles which are successfully packed have the 'was_packed' flag
// set to a non-zero value and 'x' and 'y' store the minimum location
// on each axis (i.e. bottom-left in cartesian coordinates, top-left
// if you imagine y increasing downwards). Rectangles which do not fit
// have the 'was_packed' flag set to 0.
//
// You should not try to access the 'rects' array from another thread
// while this function is running, as the function temporarily reorders
// the array while it executes.
//
// To pack into another rectangle, you need to call STBRPInitTarget
// again. To continue packing into the same rectangle, you can call
// this function again. Calling this multiple times with multiple rect
// arrays will probably produce worse packing results than calling it
// a single time with the full rectangle array, but the option is
// available.
//
// The function returns 1 if all of the rectangles were successfully
// packed and 0 otherwise.

type STBRPRect struct {
	// reserved for your use:
	id int

	// input:
	w, h STBRPCoord

	// output:
	x, y       STBRPCoord
	was_packed int // non-zero if valid packing

} // 16 bytes, nominally

func NewSTBRPRect(w, h int) *STBRPRect {
	rect := &STBRPRect{}
	rect.SetDimensions(w, h)
	return rect
}

func (rect *STBRPRect) SetDimensions(w, h int) {
	rect.w = STBRPCoord(w)
	rect.h = STBRPCoord(h)
}

// STBRP_DEF void STBRPInitTarget (stbrp_context *context, int width, int height, stbrp_node *nodes, int num_nodes);
// Initialize a rectangle packer to:
//    pack a rectangle that is 'width' by 'height' in dimensions
//    using temporary storage provided by the array 'nodes', which is 'num_nodes' long
//
// You must call this function every time you start packing into a new target.
//
// There is no "shutdown" function. The 'nodes' memory must stay valid for
// the following STBRPPackRects() call (or calls), but can be freed after
// the call (or calls) finish.
//
// Note: to guarantee best results, either:
//       1. make sure 'num_nodes' >= 'width'
//   or  2. call stbrp_allow_out_of_mem() defined below with 'allow_out_of_mem = 1'
//
// If you don't do either of the above things, widths will be quantized to multiples
// of small integers to guarantee the algorithm doesn't run out of temporary storage.
//
// If you do #2, then the non-quantized algorithm will be used, but the algorithm
// may run out of temporary storage and be unable to pack some rectangles.

// STBRP_DEF void stbrp_setup_allow_out_of_mem (stbrp_context *context, int allow_out_of_mem);
// Optionally call this function after init but before doing any packing to
// change the handling of the out-of-temp-memory scenario, described above.
// If you call init again, this will be reset to the default (false).

// STBRP_DEF void stbrp_setup_heuristic (stbrp_context *context, int heuristic);
// Optionally select which packing heuristic the library should use. Different
// heuristics will produce better/worse results for different data sets.
// If you call init again, this will be reset to the default.

const (
	STBRP_HEURISTIC_Skyline_default       = 0
	STBRP_HEURISTIC_Skyline_BL_sortHeight = 0
	STBRP_HEURISTIC_Skyline_BF_sortHeight = 1
)

//////////////////////////////////////////////////////////////////////////////
//
// the details of the following structures don't matter to you, but they must
// be visible so you can handle the memory allocations for them

type STBRPNode struct {
	x, y STBRPCoord
	next *STBRPNode
}

type STBRPContext struct {
	width       int
	height      int
	align       int
	init_mode   int
	heuristic   int
	num_nodes   int
	active_head *STBRPNode
	free_head   *STBRPNode
	extra       []STBRPNode // we allocate two extra nodes so optimal user-node-count is 'width' not 'width+2'
}

//////////////////////////////////////////////////////////////////////////////
//
//     IMPLEMENTATION SECTION
//

const STBRP__INIT_skyline = 1

func stbrp_setup_heuristic(context *STBRPContext, heuristic int) error {
	switch context.init_mode {
	case STBRP__INIT_skyline:
		if !(heuristic == STBRP_HEURISTIC_Skyline_BL_sortHeight || heuristic == STBRP_HEURISTIC_Skyline_BF_sortHeight) {
			return errors.New(fmt.Sprintf("invalid value for heuristic: %v", heuristic))
		}
		context.heuristic = heuristic
	default:
		return errors.New(fmt.Sprintf("invalid value for context.init_mode: %v", context.init_mode))
	}
	return nil
}

func stbrpSetupAllowOutOfMem(context *STBRPContext, allow_out_of_mem bool) error {
	if allow_out_of_mem {
		// if it's ok to run out of memory, then don't bother aligning them;
		// this gives better packing, but may fail due to OOM (even though
		// the rectangles easily fit). @TODO a smarter approach would be to only
		// quantize once we've hit OOM, then we could get rid of this parameter.
		context.align = 1
	} else {
		// if it's not ok to run out of memory, then quantize the widths
		// so that num_nodes is always enough nodes.
		//
		// I.e. num_nodes * align >= width
		//                  align >= width / num_nodes
		//                  align = ceil(width/num_nodes)

		context.align = (context.width + context.num_nodes - 1) / context.num_nodes
	}
	return nil
}

func STBRPInitTarget(context *STBRPContext, width, height int, nodes []*STBRPNode) {
	var i int
	for i = 0; i < len(nodes)-1; i++ {
		nodes[i].next = nodes[i+1]
	}
	nodes[i].next = nil
	context.init_mode = STBRP__INIT_skyline
	context.heuristic = STBRP_HEURISTIC_Skyline_default
	context.free_head = nodes[0]
	context.active_head = &context.extra[0]
	context.width = width
	context.height = height
	context.num_nodes = len(nodes)
	stbrpSetupAllowOutOfMem(context, false)

	// node 0 is the full width, node 1 is the sentinel (lets us not store width explicitly)
	context.extra[0].x = 0
	context.extra[0].y = 0
	context.extra[0].next = &context.extra[1]
	context.extra[1].x = STBRPCoord(width)
	context.extra[1].y = (1 << 30)
	context.extra[1].next = nil
}

// find minimum y position if it starts at x1
func stbrpSkylineFindMinY(c *STBRPContext, first *STBRPNode, x0, width int, pwaste *int) int {
	var node *STBRPNode = first
	var x1 = x0 + width
	var min_y, visited_width, waste_area int

	if first.x <= STBRPCoord(x0) {
		log.Fatalf("assertion failed: first.x <= STBRPCoord(x0)")
	}
	if node.next.x > STBRPCoord(x0) {
		log.Fatalf("assertion failed: node.next.x > STBRPCoord(x0)")
	}
	if node.x <= STBRPCoord(x0) {
		log.Fatalf("assertion failed: node.x <= STBRPCoord(x0)")
	}

	min_y = 0
	waste_area = 0
	visited_width = 0
	for node.x < STBRPCoord(x1) {
		if node.y > STBRPCoord(min_y) {
			// raise min_y higher.
			// we've accounted for all waste up to min_y,
			// but we'll now add more waste for everything we've visted
			waste_area += int(STBRPCoord(visited_width) * (node.y - STBRPCoord(min_y)))
			min_y = int(node.y)
			// the first time through, visited_width might be reduced
			if node.x < STBRPCoord(x0) {
				visited_width += int(node.next.x) - x0
			} else {
				visited_width += int(node.next.x - node.x)
			}
		} else {
			// add waste area
			var under_width = int(node.next.x - node.x)
			if under_width+visited_width > width {
				under_width = width - visited_width
			}
			waste_area += under_width * (min_y - int(node.y))
			visited_width += under_width
		}
		node = node.next
	}

	*pwaste = waste_area
	return min_y
}

type stbrp__findresult struct {
	x, y      int
	prev_link **STBRPNode
}

func stbrpSkylineFindBestPos(c *STBRPContext, width, height int) stbrp__findresult {
	var best_waste int = (1 << 30)
	var best_x int
	var best_y = (1 << 30)
	var fr stbrp__findresult
	var prev **STBRPNode
	var node *STBRPNode
	var tail *STBRPNode
	var best **STBRPNode

	// align to multiple of c->align
	width = width + c.align - 1
	width -= width % c.align
	if width%c.align == 0 {
		log.Fatal("assertion failed: width % c.align == 0")
	}

	// if it can't possibly fit, bail immediately
	if width > c.width || height > c.height {
		fr.prev_link = nil
		fr.x = 0
		fr.y = 0
		return fr
	}

	node = c.active_head
	prev = &c.active_head
	for int(node.x)+width <= c.width {
		var y, waste int
		y = stbrpSkylineFindMinY(c, node, int(node.x), width, &waste)
		if c.heuristic == STBRP_HEURISTIC_Skyline_BL_sortHeight { // actually just want to test BL
			// bottom left
			if y < best_y {
				best_y = y
				best = prev
			}
		} else {
			// best-fit
			if y+height <= c.height {
				// can only use it if it first vertically
				if y < best_y || (y == best_y && waste < best_waste) {
					best_y = y
					best_waste = waste
					best = prev
				}
			}
		}
		prev = &node.next
		node = node.next
	}
	if best == nil {
		best_x = 0
	} else {
		best_x = int((*best).x)
	}

	// if doing best-fit (BF), we also have to try aligning right edge to each node position
	//
	// e.g, if fitting
	//
	//     ____________________
	//    |____________________|
	//
	//            into
	//
	//   |                         |
	//   |             ____________|
	//   |____________|
	//
	// then right-aligned reduces waste, but bottom-left BL is always chooses left-aligned
	//
	// This makes BF take about 2x the time

	if c.heuristic == STBRP_HEURISTIC_Skyline_BF_sortHeight {
		tail = c.active_head
		node = c.active_head
		prev = &c.active_head
		// find first node that's admissible
		for int(tail.x) < width {
			tail = tail.next
		}
		for tail != nil {
			var xpos = int(tail.x) - width
			var y, waste int
			if xpos >= 0 {
				log.Fatal("assertion failed: xpos >= 0")
			}
			// find the left position that matches this
			for int(node.next.x) <= xpos {
				prev = &node.next
				node = node.next
			}
			if int(node.next.x) > xpos && int(node.x) <= xpos {
				log.Fatal("assertion failed: node->next->x > xpos && node->x <= xpos")
			}
			y = stbrpSkylineFindMinY(c, node, xpos, width, &waste)
			if y+height <= c.height {
				if y <= best_y {
					if y < best_y || waste < best_waste || (waste == best_waste && xpos < best_x) {
						best_x = xpos
						if y <= best_y {
							log.Fatal("assertion failed: y <= best_y")
						}
						best_y = y
						best_waste = waste
						best = prev
					}
				}
			}
			tail = tail.next
		}
	}

	fr.prev_link = best
	fr.x = best_x
	fr.y = best_y
	return fr
}

func stbrpSkylinePackRectangle(context *STBRPContext, width, height int) stbrp__findresult {
	// find best position according to heuristic
	var res = stbrpSkylineFindBestPos(context, width, height)
	var node, cur *STBRPNode

	// bail if:
	//    1. it failed
	//    2. the best node doesn't fit (we don't always check this)
	//    3. we're out of memory
	if res.prev_link == nil || res.y+height > context.height || context.free_head == nil {
		res.prev_link = nil
		return res
	}

	// on success, create new node
	node = context.free_head
	node.x = STBRPCoord(res.x)
	node.y = STBRPCoord(res.y + height)

	context.free_head = node.next

	// insert the new node into the right starting point, and
	// let 'cur' point to the remaining nodes needing to be
	// stiched back in

	cur = *res.prev_link
	if int(cur.x) < res.x {
		// preserve the existing one, so start testing with the next one
		var next = cur.next
		cur.next = node
		cur = next
	} else {
		*res.prev_link = node
	}

	// from here, traverse cur and free the nodes, until we get to one
	// that shouldn't be freed
	for cur.next != nil && int(cur.next.x) <= res.x+width {
		var next = cur.next
		// move the current node to the free list
		cur.next = context.free_head
		context.free_head = cur
		cur = next
	}

	// stitch the list back in
	node.next = cur

	if int(cur.x) < res.x+width {
		cur.x = STBRPCoord(res.x + width)
	}

	return res
}

func rectHeightCompare(p, q *STBRPRect) int {
	if p.h > q.h {
		return -1
	}
	if p.h < q.h {
		return 1
	}
	if p.w > q.w {
		return -1
	} else {
		if p.w < q.w {
			return 1
		} else {
			return 0
		}
	}
}

func rectOriginalOrder(p, q *STBRPRect) int {
	if p.was_packed < q.was_packed {
		return -1
	} else {
		if p.was_packed > q.was_packed {
			return 1
		} else {
			return 0
		}
	}
}

func STBRPPackRects(context *STBRPContext, rects []*STBRPRect) int {
	var i int
	var all_rects_packed = 1

	// we use the 'was_packed' field internally to allow sorting/unsorting
	for i = 0; i < len(rects); i++ {
		rects[i].was_packed = i
	}

	// sort according to heuristic
	//qsort(rects, num_rects, sizeof(rects[0]), rect_height_compare);
	sort.Slice(rects, func(i, j int) bool { return rectHeightCompare(rects[i], rects[j]) > 0 })

	for i = 0; i < len(rects); i++ {
		if rects[i].w == 0 || rects[i].h == 0 {
			rects[i].x = 0
			rects[i].y = 0 // empty rect needs no space
		} else {
			var fr = stbrpSkylinePackRectangle(context, int(rects[i].w), int(rects[i].h))
			if fr.prev_link != nil {
				rects[i].x = STBRPCoord(fr.x)
				rects[i].y = STBRPCoord(fr.y)
			} else {
				rects[i].x = STBRP__MAXVAL
				rects[i].y = STBRP__MAXVAL
			}
		}
	}

	// unsort
	//qsort(rects, num_rects, sizeof(rects[0]), rect_original_order);
	sort.Slice(rects, func(i, j int) bool { return rectOriginalOrder(rects[i], rects[j]) > 0 })

	// set was_packed flags and all_rects_packed status
	for i = 0; i < len(rects); i++ {
		if !(int(rects[i].x) == STBRP__MAXVAL && int(rects[i].y) == STBRP__MAXVAL) {
			rects[i].was_packed = 1
		} else {
			rects[i].was_packed = 0
		}
		if rects[i].was_packed == 0 {
			all_rects_packed = 0
		}
	}

	// return the all_rects_packed status
	return all_rects_packed
}

/*
------------------------------------------------------------------------------
This software is available under 2 licenses -- choose whichever you prefer.
------------------------------------------------------------------------------
ALTERNATIVE A - MIT License
Copyright (c) 2017 Sean Barrett
Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
------------------------------------------------------------------------------
ALTERNATIVE B - Public Domain (www.unlicense.org)
This is free and unencumbered software released into the public domain.
Anyone is free to copy, modify, publish, use, compile, sell, or distribute this
software, either in source code form or as a compiled binary, for any purpose,
commercial or non-commercial, and by any means.
In jurisdictions that recognize copyright laws, the author or authors of this
software dedicate any and all copyright interest in the software to the public
domain. We make this dedication for the benefit of the public at large and to
the detriment of our heirs and successors. We intend this dedication to be an
overt act of relinquishment in perpetuity of all present and future rights to
this software under copyright law.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
------------------------------------------------------------------------------
*/
