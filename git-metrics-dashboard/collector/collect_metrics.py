import os
import sys
from collections import defaultdict
from datetime import datetime
from prometheus_client import CollectorRegistry, Gauge, push_to_gateway
from status_analyzer import LeadTimeAnalyzer

REPO_OWNER = "SVIN-Team"
REPO_NAME = "bmstu_student_app"
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN", "")
PUSHGATEWAY_URL = os.getenv("PUSHGATEWAY_URL", "pushgateway:9091")

if not GITHUB_TOKEN:
    print(" Ошибка: GITHUB_TOKEN не установлен")
    sys.exit(1)

def export_metrics():
    print("ЗАПУСК СБОРА МЕТРИК ИЗ GITHUB")
    print(f" Репозиторий: {REPO_OWNER}/{REPO_NAME}")
    print(f" Время: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    analyzer = LeadTimeAnalyzer(REPO_OWNER, REPO_NAME, GITHUB_TOKEN)
    
    print("\n Анализируем завершенные элементы (с статусом 'Done')...")
    monthly_data = analyzer.analyze_lead_time_by_month()
    
    if not monthly_data:
        print(" Нет данных о завершенных элементах!")
        print("   Возможные причины:")
        print("   1. Нет элементов со статусом 'Done'")
        print("   2. Ошибка при получении данных из GitHub")
        print("   3. Проект не найден")
        return
    
    registry = CollectorRegistry()
    
    lead_time_gauge = Gauge(
        "lead_time_status_days",
        "Lead time by month and status (days)",
        ["month", "status"],
        registry=registry
    )
    
    lead_time_gauge_hours = Gauge(
        "lead_time_status_hours",
        "Lead time by month and status (hours)",
        ["month", "status"],
        registry=registry
    )
    
    print("\n МЕТРИКИ ПО МЕСЯЦАМ ДЛЯ STACKED BAR CHART:")
    
    all_metrics = []
    
    for month, status_times in sorted(monthly_data.items()):
        print(f"\n {month}:")
        month_total = 0
        
        for status in analyzer.statuses_order:
            times = status_times.get(status, [])
            if times:
                avg_days = sum(times) / len(times)
                avg_hours = avg_days * 24
                month_total += avg_days
                
                status_name = analyzer.status_english[status]
                
                lead_time_gauge.labels(month=month, status=status_name).set(avg_days)
                lead_time_gauge_hours.labels(month=month, status=status_name).set(avg_hours)
                
                display_name = analyzer.status_display[status]
                print(f"   {display_name}: {avg_days:.2f} дн ({avg_hours:.1f} ч) - на основе {len(times)} элементов")
                
                all_metrics.append({
                    "month": month,
                    "status": display_name,
                    "days": avg_days,
                    "hours": avg_hours,
                    "count": len(times)
                })
        
        print(f" Итого в месяце: {month_total:.2f} дней")
    
    total_lead_time_gauge = Gauge(
        "total_lead_time_days",
        "Total lead time by month (days)",
        ["month"],
        registry=registry
    )
    
    for month, status_times in sorted(monthly_data.items()):
        month_total = 0
        for status in analyzer.statuses_order:
            times = status_times.get(status, [])
            if times:
                month_total += sum(times) / len(times)
        total_lead_time_gauge.labels(month=month).set(month_total)
    
    total_completed = 0
    for month, status_times in monthly_data.items():
        for status, times in status_times.items():
            total_completed += len(times)
    
    completed_gauge = Gauge("completed_items_count", "Number of completed items", registry=registry)
    completed_gauge.set(total_completed)
    print(f"\n Всего переходов по статусам: {total_completed}")
    
    print(f" Отправка метрик в Pushgateway: {PUSHGATEWAY_URL}")
    
    try:
        push_to_gateway(PUSHGATEWAY_URL, job="git_lead_time", registry=registry)
        print(f" Метрики отправлены в {PUSHGATEWAY_URL}")
        
        print("\n   Пример метрик:")
        for metric in all_metrics[:5]:
            print(f"     {metric['month']} - {metric['status']}: {metric['days']:.2f} дней")
        if len(all_metrics) > 5:
            print(f"     ... и еще {len(all_metrics)-5} метрик")
            
    except Exception as e:
        print(f" ОШИБКА отправки: {e}")
        sys.exit(1)
    
    print(" РАБОТА ЗАВЕРШЕНА")

if __name__ == "__main__":
    export_metrics()