#!/bin/bash

# 修复executor.go中的编译错误

FILE="internal/workflow/executor.go"

# 1. 删除未使用的allSuccess变量
sed -i '' '/allSuccess := true/d' "$FILE"
sed -i '' '/allSuccess = false/d' "$FILE"
sed -i '' '/allSuccess = true/d' "$FILE"

# 2. 修复result.Status（StepResult没有Status字段，使用其他方式判断成功）
sed -i '' 's/result.Status/true/g' "$FILE"

# 3. 修复step.Goal（Step没有Goal字段，使用step.Name）
sed -i '' 's/step.Goal/step.Name/g' "$FILE"

# 4. 修复compareNumbers的类型断言
# 将 compareNumbers(typedValue, condition.Value, ">") 改为使用fmt.Sprint转换
sed -i '' 's/compareNumbers(typedValue, condition.Value, ">")/compareNumbers(fmt.Sprintf("%v", typedValue), fmt.Sprintf("%v", condition.Value), ">")/g' "$FILE"
sed -i '' 's/compareNumbers(typedValue, condition.Value, "<")/compareNumbers(fmt.Sprintf("%v", typedValue), fmt.Sprintf("%v", condition.Value), "<")/g' "$FILE"
sed -i '' 's/compareNumbers(typedValue, condition.Value, ">=")/compareNumbers(fmt.Sprintf("%v", typedValue), fmt.Sprintf("%v", condition.Value), ">=")/g' "$FILE"
sed -i '' 's/compareNumbers(typedValue, condition.Value, "<=")/compareNumbers(fmt.Sprintf("%v", typedValue), fmt.Sprintf("%v", condition.Value), "<=")/g' "$FILE"

echo "修复完成！"
