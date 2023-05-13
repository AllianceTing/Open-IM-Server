// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"Open_IM/internal/cron_task"
	"flag"
	"fmt"
	"time"
)

func main() {
	var userID = flag.String("userID", "", "userID to clear msg and reset seq")
	var workingGroupID = flag.String("workingGroupID", "", "workingGroupID to clear msg and reset seq")
	flag.Parse()
	fmt.Println(time.Now(), "start cronTask", *userID, *workingGroupID)
	cronTask.StartCronTask(*userID, *workingGroupID)
}
