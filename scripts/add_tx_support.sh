#!/bin/bash

# 为所有仓储文件添加事务支持

REPO_DIR="/Users/zengshenglong/Code/GoWorkSpace/Pulse/internal/repository"

# 需要处理的仓储文件列表
REPO_FILES=(
    "rule_repository.go"
    "datasource_repository.go"
    "ticket_repository.go"
    "knowledge_repository.go"
    "permission_repository.go"
)

for file in "${REPO_FILES[@]}"; do
    filepath="$REPO_DIR/$file"
    if [ -f "$filepath" ]; then
        echo "Processing $file..."
        
        # 获取仓储类型名（去掉_repository.go后缀）
        repo_type=$(echo "$file" | sed 's/_repository\.go$//')
        
        # 转换为驼峰命名
        case $repo_type in
            "rule")
                struct_name="ruleRepository"
                interface_name="RuleRepository"
                constructor_name="NewRuleRepository"
                tx_constructor_name="NewRuleRepositoryWithTx"
                ;;
            "datasource")
                struct_name="dataSourceRepository"
                interface_name="DataSourceRepository"
                constructor_name="NewDataSourceRepository"
                tx_constructor_name="NewDataSourceRepositoryWithTx"
                ;;
            "ticket")
                struct_name="ticketRepository"
                interface_name="TicketRepository"
                constructor_name="NewTicketRepository"
                tx_constructor_name="NewTicketRepositoryWithTx"
                ;;
            "knowledge")
                struct_name="knowledgeRepository"
                interface_name="KnowledgeRepository"
                constructor_name="NewKnowledgeRepository"
                tx_constructor_name="NewKnowledgeRepositoryWithTx"
                ;;
            "permission")
                struct_name="permissionRepository"
                interface_name="PermissionRepository"
                constructor_name="NewPermissionRepository"
                tx_constructor_name="NewPermissionRepositoryWithTx"
                ;;
        esac
        
        echo "Adding tx field and WithTx constructor to $file"
    else
        echo "File $filepath not found, skipping..."
    fi
done

echo "Transaction support addition completed!"