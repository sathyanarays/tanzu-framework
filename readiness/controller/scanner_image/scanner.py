from subprocess import check_output
import json
import os

print("Scan started")
print(os.environ['DEPLOYMENT_NAME'])
print(os.environ['DEPLOYMENT_NAMESPACE'])
deployment = check_output(["kubectl", "get", "deployments", os.environ['DEPLOYMENT_NAME'], "-n", os.environ['DEPLOYMENT_NAMESPACE'], "-o", "json"])
deployment_json = json.loads(deployment)
containers_list = deployment_json["spec"]["template"]["spec"]["containers"]

target_severity = os.environ['SEVERITY']

for container in containers_list:
    image = container["image"]    
    scan_result_string = check_output(["trivy", "image", "-f", "json",image])
    scan_result_json = json.loads(scan_result_string)
    for result in scan_result_json['Results']:
        if "Vulnerabilities" not in result:
            continue
        for vuln in result['Vulnerabilities']:
            if target_severity == "LOW":
                if vuln['Severity'] == "LOW" or vuln['Severity'] == "HIGH" or vuln['Severity'] == "CRITICAL":
                    exit(1)
            if target_severity == "HIGH":
                if vuln['Severity'] == "HIGH" or vuln['Severity'] == "CRITICAL":
                    exit(1)
            if target_severity == "CRITICAL":
                if vuln['Severity'] == "CRITICAL":
                    exit(1)
print("Scan successfull")




