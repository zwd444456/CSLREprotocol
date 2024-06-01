import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

# 读取 Excel 文件中的数据
data = pd.read_excel("62PC的正常执行的延迟的吞吐量.xlsx")

# 分片成员数量的列表
x = data['交易轮数']
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


plt.plot(x, data['正常1'], label='SYNC', linestyle='-', marker='o', color=colors['SYNC'])  # 蓝色

plt.plot(x, data['正常2'], label='RBSMR', linestyle='-', marker='o', color=colors['RBSMR'])  # 紫色


# 设置 x 和 y 轴的标签
plt.xlabel('Transcation Rounds', fontsize=18)  # 分片成员数量
plt.ylabel('Throughput(bytes/s)', fontsize=18)  # 延迟（毫秒）
plt.xticks(fontsize=13)
plt.yticks(fontsize=13)
# 添加图例，确保没有重复的图例
plt.legend(ncol=2, fontsize=13)
for spine in plt.gca().spines.values():
    spine.set_edgecolor('black')
    spine.set_linewidth(2)

# 设置 x 和 y 轴的范围


# 显示图形
plt.show()
import numpy as np
import numpy as np
# import pandas as pd
# import matplotlib.pyplot as plt
# import numpy as np
#
# # 使用数据直接创建 DataFrame
# data = pd.DataFrame({
#     '交易轮数': [1, 5, 10, 20, 50, 100, 200],
#     '正常1': [668.2884, 3400.442, 6882.884, 13785.768, 34514.42, 68888.84, 137957.68],
#     '正常2': [968.2884, 4862.442, 9616.884, 19380.768, 49106.42, 98500.84, 197667.68]
# })
#
# # 获取横轴（交易轮数）和每个协议的值
# x = np.array(data['交易轮数'])  # 交易轮数
# bar_width = 0.3  # 每个柱状图的宽度
# x_pos = np.arange(len(x))  # 获取x轴的位置
#
# # 创建图形
# fig, ax = plt.subplots(figsize=(10, 6))
#
# # 绘制不同的柱状图
# ax.bar(x_pos - bar_width / 2, data['正常1'], width=bar_width, label='SYNC', color='#8fd3c8')  # 蓝色
# ax.bar(x_pos + bar_width / 2, data['正常2'], width=bar_width, label='RBSMR', color='#d9d9d9')  # 橙色
#
# # 设置标签
# ax.set_xlabel('transcation rounds', fontsize=18)
# ax.set_ylabel('Throughput(byte/s)', fontsize=18)
#
#
# # 设置x轴的刻度值
# ax.set_xticks(x_pos)
# ax.set_xticklabels(x)
#
# # 添加图例
# ax.legend()
#
# # 显示图形
# plt.tight_layout()
# plt.show()
