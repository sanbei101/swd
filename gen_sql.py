import os

data_dir = 'sensitive_words'
output_file = 'data.sql'
table_name = 'sensitive_words'
batch_size = 10000

def generate_sql():
    if not os.path.exists(data_dir):
        print(f"错误:目录 {data_dir} 不存在")
        return

    all_data = []
    
    for filename in os.listdir(data_dir):
        if filename.endswith('.txt'):
            word_type = filename.replace('.txt', '')
            file_path = os.path.join(data_dir, filename)
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    for line in f:
                        word = line.strip()
                        if word:
                            safe_word = word.replace("'", "''")
                            # 存储为元组 (word, type)
                            all_data.append((safe_word, word_type))
            except Exception as e:
                print(f"读取 {filename} 出错: {e}")

    with open(output_file, 'w', encoding='utf-8') as f:
        # 建表头
        f.write(f"CREATE TABLE IF NOT EXISTS {table_name} (\n")
        f.write("    id SERIAL PRIMARY KEY,\n")
        f.write("    word varchar(255) NOT NULL,\n")
        f.write("    type varchar(50) NOT NULL DEFAULT 'default'\n")
        f.write(");\n\n")
        
        f.write("BEGIN;\n")
        
        for i in range(0, len(all_data), batch_size):
            batch = all_data[i : i + batch_size]
            values_str = ",\n".join([f"('{w}', '{t}')" for w, t in batch])
            sql = f"INSERT INTO {table_name} (word, type) VALUES \n{values_str};\n"
            f.write(sql)
            
        f.write("COMMIT;\n")

    print(f"处理完成!总计 {len(all_data)} 条词汇，已保存至 {output_file}")

if __name__ == "__main__":
    generate_sql()