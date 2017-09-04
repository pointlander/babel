# Copyright 2017 The Babel Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

profile:
	go tool pprof -text -lines babel /tmp/profile
