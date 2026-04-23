#!/usr/bin/env python3

from datetime import datetime, timezone
from collections import defaultdict
import requests
import os

class LeadTimeAnalyzer:
    def __init__(self, repo_owner, repo_name, github_token):
        self.repo_owner = repo_owner
        self.repo_name = repo_name
        self.graphql_headers = {
            "Authorization": f"Bearer {github_token}",
            "Content-Type": "application/json"
        }
        
        self.statuses_order = [
            "Backlog", "Ready", "In progress", "In review",
            "Pre-testing", "Checking", "Testing", "Upload required", "Done"
        ]
        
        self.status_display = {
            "Backlog": "Бэклог",
            "Ready": "Готово к работе",
            "In progress": "В работе",
            "In review": "Ревью",
            "Pre-testing": "Подготовка к тестированию",
            "Checking": "На проверке",
            "Testing": "Тестируется",
            "Upload required": "Требуется выгрузка",
            "Done": "Готово"
        }
        
        self.status_english = {
            "Backlog": "backlog",
            "Ready": "ready",
            "In progress": "in_progress",
            "In review": "in_review",
            "Pre-testing": "pre_testing",
            "Checking": "checking",
            "Testing": "testing",
            "Upload required": "upload_required",
            "Done": "done"
        }
    
    def get_all_issues(self):
        url = f"https://api.github.com/repos/{self.repo_owner}/{self.repo_name}/issues"
        params = {"state": "all", "per_page": 100, "filter": "all"}
        
        all_issues = []
        page = 1
        
        while True:
            params["page"] = page
            response = requests.get(url, headers={"Authorization": f"token {os.getenv('GITHUB_TOKEN')}"}, params=params)
            if response.status_code != 200:
                break
            
            issues = response.json()
            if not issues:
                break
            
            for issue in issues:
                if 'pull_request' not in issue:
                    all_issues.append(issue)
            
            page += 1
            if "next" not in response.links:
                break
        
        return all_issues
    
    def get_issue_status_history(self, issue_number):
        query = """
        query($owner: String!, $repo: String!, $issueNumber: Int!) {
          repository(owner: $owner, name: $repo) {
            issue(number: $issueNumber) {
              timelineItems(first: 50, itemTypes: [PROJECT_V2_ITEM_STATUS_CHANGED_EVENT]) {
                nodes {
                  ... on ProjectV2ItemStatusChangedEvent {
                    createdAt
                    status
                  }
                }
              }
            }
          }
        }
        """
        
        variables = {
            "owner": self.repo_owner,
            "repo": self.repo_name,
            "issueNumber": issue_number
        }
        
        response = requests.post(
            "https://api.github.com/graphql",
            headers=self.graphql_headers,
            json={"query": query, "variables": variables}
        )
        
        if response.status_code != 200:
            return []
        
        data = response.json()
        
        if "errors" in data:
            return []
        
        timeline = data.get("data", {}).get("repository", {}).get("issue", {}).get("timelineItems", {}).get("nodes", [])
        
        status_changes = []
        for event in timeline:
            if event and event.get('status'):
                created_at = datetime.fromisoformat(event['createdAt'].replace('Z', '+00:00'))
                status_changes.append({
                    'status': event['status'],
                    'time': created_at
                })
        
        return status_changes
    
    def analyze_lead_time_by_month(self):
        print("АНАЛИЗ LEAD TIME")
        
        issues = self.get_all_issues()
        print(f"  Найдено issues: {len(issues)}")
        
        monthly_data = defaultdict(lambda: defaultdict(list))
        now = datetime.now(timezone.utc)
        
        for issue in issues:
            issue_number = issue['number']
            issue_title = issue['title'][:40]
            
            print(f"\n--- #{issue_number}: {issue_title} ---")
            
            status_changes = self.get_issue_status_history(issue_number)
            
            if not status_changes:
                print(f"  Нет истории статусов")
                continue
            
            for change in status_changes:
                print(f" {change['time'].strftime('%Y-%m-%d %H:%M:%S')} -> {change['status']}")
            
            status_changes.sort(key=lambda x: x['time'])
            
            last_time = status_changes[-1]['time']
            month_key = last_time.strftime("%Y-%m")
            
            print(f" Группировка в {month_key} (последнее изменение)")
            
            for i in range(len(status_changes)):
                current = status_changes[i]
                
                if i + 1 < len(status_changes):
                    next_time = status_changes[i + 1]['time']
                else:
                    next_time = now
                
                duration_seconds = (next_time - current['time']).total_seconds()
                duration_days = duration_seconds / 86400
                
                if duration_days > 0:
                    monthly_data[month_key][current['status']].append(duration_days)
                    print(f"     {current['status']}: {duration_days:.2f} дн (до {next_time.strftime('%Y-%m-%d %H:%M:%S')})")
        
        return monthly_data
    
    def push_metrics_to_pushgateway(self, pushgateway_url):
        print("\n Отправка метрик...")
        
        monthly_data = self.analyze_lead_time_by_month()
        
        if not monthly_data:
            print(" Нет данных")
            return
        
        all_metrics = []
        
        sorted_months = sorted(monthly_data.keys())
        
        for month in sorted_months:
            status_times = monthly_data[month]
            print(f"\n {month}:")
            for status in self.statuses_order:
                times = status_times.get(status, [])
                if times:
                    avg_days = sum(times) / len(times)
                    metric = f'lead_time_status_days{{month="{month}",status="{self.status_english[status]}"}} {avg_days:.2f}'
                    all_metrics.append(metric)
                    print(f"     {self.status_display[status]}: {avg_days:.2f} дн ({len(times)} элементов)")
        
        if all_metrics:
            payload = "\n".join(all_metrics)
            url = f"{pushgateway_url}/metrics/job/git_lead_time"
            
            try:
                requests.delete(url, timeout=5)
            except:
                pass
            
            response = requests.post(url, data=payload, headers={"Content-Type": "text/plain"})
            if response.status_code == 202:
                print(f"\n Отправлено {len(all_metrics)} метрик")
            else:
                print(f" Ошибка: {response.status_code}")

def export_metrics():
    print("\n ЗАПУСК СБОРА МЕТРИК")
    
    github_token = os.getenv("GITHUB_TOKEN")
    pushgateway_url = os.getenv("PUSHGATEWAY_URL", "http://pushgateway:9091")
    
    if not github_token:
        print(" Нет GITHUB_TOKEN")
        return
    
    analyzer = LeadTimeAnalyzer("SVIN-Team", "bmstu_student_app", github_token)
    analyzer.push_metrics_to_pushgateway(pushgateway_url)

if __name__ == "__main__":
    export_metrics()