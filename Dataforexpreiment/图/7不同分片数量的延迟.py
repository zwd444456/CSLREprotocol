import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from matplotlib.font_manager import FontProperties

# Read data from Excel file
data = pd.read_excel('不同分片数量的时延.xlsx')

# Get x-axis values and y-axis values for each protocol
x = np.array(data['分片数量'])  # Transaction rounds (x-axis)
protocols = {
    'SYNC': data['正常1'],
  'SYNC-CSVC': data['协议11'],
  'SYNC-FS': data['协议12'],
  'SYNC-CSLRE': data['协议13'],
    'RBSMR': data['正常2'],
   'RBSMR-CSVC': data['协议21'],
   'RBSMR-FS': data['协议22'],
   'RBSMR-CSLRE': data['协议23']
}

# Bar width and position for the x-axis
bar_width = 0.12
x_pos = np.arange(len(x))  # Position of the x-axis

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

# Plotting each protocol with offset for clarity
fig, ax = plt.subplots(figsize=(12, 8))
for i, (protocol_name, values) in enumerate(protocols.items()):
    ax.bar(x_pos + i * bar_width, values, bar_width, label=protocol_name, color=colors[protocol_name])

# Labels and title
ax.set_xlabel('Shard members', fontsize=30)  # 增加 x 轴标签字体大小
ax.set_ylabel('Latency (ms)', fontsize=25)  # 增加 y 轴标签字体大小
plt.xticks(fontsize=25)
plt.yticks(fontsize=22)
plt.legend(ncol=2, fontsize=22)
for spine in plt.gca().spines.values():
    spine.set_edgecolor('black')
    spine.set_linewidth(2)
# Ticks and legend
ax.set_xticks(x_pos + (bar_width * (len(protocols) - 1) / 2))  # Centering the x-axis ticks
ax.set_xticklabels(x)  # Labeling the x-axis
ax.set_ylim(0, max(data.max()) * 1.2)  # Increase y-axis limit to 120% of the maximum value


# Display the plot
plt.show()
