#!/bin/bash

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# if [[ -z "${GATEWAY_URL:-}" ]];then
#   GATEWAY_URL=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}' || true)
# fi
# 
# HTTPS=${HTTPS:-"false"}
# H2UPGRADE=${H2UPGRADE:-"false"}

function run_test() {
  local ns=${1:?"namespaces"}
  local prefix=${2:?"prefix name for service. typically svc-"}
  local manifestDir=${3:?"path to the manifest directory"}

  # YAML=$(mktemp).yml
  # # shellcheck disable=SC2086
  # helm -n ${ns} template \
  #         --set serviceNamePrefix="${prefix}" \
  #         --set Namespace="${ns}" \
  #         --set domain="${DNS_DOMAIN}" \
  #         --set ingress="${GATEWAY_URL}" \
  #         --set https="${HTTPS}" \
  #         --set h2upgrade="${H2UPGRADE}" \
  #         . > "${YAML}"
  # echo "Wrote ${YAML}"

  # kubectl create ns "${ns}" || true
  # kubectl label namespace "${ns}" "${INJECTION_LABEL:-istio-injection=enabled}" --overwrite

  if [[ -z "${DELETE}" ]];then
    sleep 3
    for manifest in ${manifestDir}/*.yaml; do
       kubectl apply -f "${manifest}" --context "$(echo ${manifest} | cut -d'.' -f1 | cut -d'/' -f2)" &
    done
    wait
  else
    for manifest in ${manifestDir}/*.yaml; do
      kubectl delete -f "${manifest}" --context "$(echo ${manifest} | cut -d'.' -f1 | cut -d'/' -f2)" &
    done
    wait
    # kubectl delete ns "${ns}"
  fi

  echo "Done with run tests"
}

function start_servicegraphs() {
  local nn=${1:?"number of namespaces"}
  local min=${2:-"0"}
  local manifestDir=${3:?"path to manifest directory"}

   # shellcheck disable=SC2004
   for ((ii=$min; ii<$nn; ii++)) {
    ns=$(printf 'service-graph%.2d' $ii)
    prefix=$(printf 'svc%.2d-' $ii)
    if [[ -z "${DELETE}" ]];then
      ${CMD} run_test "${ns}" "${prefix}" "${manifestDir}"
      # ${CMD} "${WD}/loadclient/setup_test.sh" "${ns}" "${prefix}"
    else
      # ${CMD} "${WD}/loadclient/setup_test.sh" "${ns}" "${prefix}"
      ${CMD} run_test "${ns}" "${prefix}" "${manifestDir}"
    fi

    sleep 30
  }
}
