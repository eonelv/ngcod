/*技能{施法对象, 释放目标, 技能类型}
1. 收到客户端消息后，转发给对应的user处理
2. user发现是技能消息， 转给技能消息处理
3. 根据技能类型调用对应技能的逻辑
4. 判断技能满足释放条件
5. 筛选技能作用对象（对象在哪里？地图块中？）
6. 调用伤害计算函数，计算每个对象的伤害[伤害计算公式是统一的]
7. 更新每个对象属性及buff
8. 将技能施法结果广播给同屏的人（技能类型，施法着，作用对象，每个对象打出的伤害）
*/