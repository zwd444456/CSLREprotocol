import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

# 读取 Excel 文件中的数据
data = pd.read_excel("62PC的正常执行的延迟的时间复杂度.xlsx")

# 分片成员数量的列表
x = data['交易轮数']

# 定义不同线条的颜色
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


# 将 y 轴数值除以 1000
data['正常1'] = data['正常1'] / 1000
data['正常2'] = data['正常2'] / 1000

# 绘制不同的延迟曲线
plt.plot(x, data['正常1'], label='SYNC', linestyle='-', marker='o', color=colors['SYNC'])  # 蓝色
plt.plot(x, data['正常2'], label='RBSMR', linestyle='-', marker='o', color=colors['RBSMR'])  # 紫色

# 设置 x 和 y 轴的标签
plt.xlabel('Transaction Rounds', fontsize=18)
plt.ylabel('Latency (s)', fontsize=18)
plt.xticks(fontsize=13)
plt.yticks(fontsize=13)
# 设置边框颜色和宽度
for spine in plt.gca().spines.values():
    spine.set_edgecolor('black')
    spine.set_linewidth(2)

# 添加图例并确保没有重复的图例
plt.legend(ncol=2, fontsize=15)

# 显示图形
plt.show()
