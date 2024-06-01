import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

# 读取 Excel 文件中的数据
data = pd.read_excel("4最大传输时延变化的时间复杂度.xlsx")

# 分片成员数量的列表
x = data['Δ']
colors = {
    'SYNC': '#045275',  # 蓝色
    'SYNC-CSVC': '#089099',  # 橙色
    'SYNC-FS': '#7CCBA2',  # 绿色
    'SYNC-CSLRE': '#D9C27A',  # 红色
    'RBSMR': '#F0746E',  # 紫色
    'RBSMR-CSVC': '#DC3977',  # 棕色
    'RBSMR-FS': '#7C1D6F',  # 粉红色
    'RBSMR-CSLRE': '#8baa55'  # 灰色
}

plt.figure(figsize=(12, 8))
# 绘制不同的延迟曲线，增大标记尺寸
plt.plot(x, data['正常1'], label='SYNC', linestyle='-', marker='p', color=colors['SYNC'], markersize=10)  # 蓝色
plt.plot(x, data['协议11'], label='SYNC-CSVC', linestyle='--', marker='D', color=colors['SYNC-CSVC'], markersize=10)  # 橙色
plt.plot(x, data['协议12'], label='SYNC-FS', linestyle='--', marker='s', color=colors['SYNC-FS'], markersize=10)  # 绿色
plt.plot(x, data['协议13'], label='SYNC-CSLRE', marker='^', color=colors['SYNC-CSLRE'], markersize=10)  # 红色
plt.plot(x, data['正常2'], label='RBSMR', linestyle='-', marker='p', color=colors['RBSMR'], markersize=10)  # 紫色
plt.plot(x, data['协议21'], label='RBSMR-CSVC', linestyle='--', marker='D', color=colors['RBSMR-CSVC'], markersize=10)  # 棕色
plt.plot(x, data['协议22'], label='RBSMR-FS', linestyle='--', marker='s', color=colors['RBSMR-FS'], markersize=10)  # 粉红色
plt.plot(x, data['协议23'], label='RBSMR-CSLRE', marker='^', color=colors['RBSMR-CSLRE'], markersize=10)  # 灰色
 # 灰色
# 设置 x 和 y 轴的标签
plt.xlabel('Maximum Network Delay ∆ (ms)',fontsize=25)  # 分片成员数量
plt.ylabel('Latency (s)',fontsize=25)  # 延迟（毫秒）



# 添加图例，确保没有重复的图例
# 设置刻度字体大小
plt.xticks(fontsize=25)
plt.yticks(fontsize=22)

# 添加图例，并增加字体大小

# 添加图例，并增加字体大小
plt.legend(ncol=2, fontsize=22)
for spine in plt.gca().spines.values():
    spine.set_edgecolor('black')
    spine.set_linewidth(2)

# 显示图形
plt.show()
