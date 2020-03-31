### Copyright (c) 2020 Richard Youngkin. All rights reserved.
### Use of this source code is governed by a MIT-style
### license that can be found in the LICENSE file.

FROM alpine

RUN apk add --no-cache curl

COPY ./todod /bin/todod
RUN ["chmod", "+x", "/bin/todod"]
ENTRYPOINT ["/bin/todod"]