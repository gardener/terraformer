# Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM alpine:3.7

ENV TF_VERSION=0.11.6
ENV TF_DEV=true
ENV TF_RELEASE=true
ENV ZONEINFO=/zone-info/zoneinfo.zip 

WORKDIR /

RUN apk add --update curl bash && \
    curl https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_linux_amd64.zip > terraform_${TF_VERSION}_linux_amd64.zip && \
    unzip terraform_${TF_VERSION}_linux_amd64.zip -d /bin && \
    rm -f terraform_${TF_VERSION}_linux_amd64.zip

ADD ./assets/zoneinfo.zip /zone-info/zoneinfo.zip
ADD ./terraform.sh /terraform.sh
ADD ./terraform-provider* /terraform-providers/

CMD exec /terraform.sh
