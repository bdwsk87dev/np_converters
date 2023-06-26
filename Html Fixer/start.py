import xml.etree.ElementTree as ET
from bs4 import BeautifulSoup
import lxml.html
import html

def unescape_special_characters(text):
    unescaped_text = html.unescape(text)
    return unescaped_text

def validate_and_fix_html(html_string):
    # Исправление HTML с помощью BeautifulSoup

    # Создание объекта BeautifulSoup для парсинга HTML
    soup = BeautifulSoup(html_string, 'lxml')

    # Удаление пустых тегов
    for tag in soup.find_all():
        if not tag.contents:
            tag.extract()

    # Исправление неправильно закрытых тегов
    soup = soup.prettify()

    # Валидация и исправление HTML с помощью lxml.html
    try:
        document = lxml.html.fromstring(str(soup))
        repaired_html = lxml.html.tostring(document, encoding='unicode')
        return repaired_html
    except:
        # Обработка ошибок валидации
        return None

# Открытие XML-файла
xml_file_path = 'file.xml'

tree = ET.parse(xml_file_path)
root = tree.getroot()

# Поиск элементов description и description_ua
description_elements = root.findall(".//description")
description_ua_elements = root.findall(".//description_ua")

# Обработка значений внутри элементов
for element in description_elements + description_ua_elements:
    cdata_value = element.text.strip()
    repaired_html = validate_and_fix_html(cdata_value)

    if repaired_html is not None:
        # Замена содержимого элемента на исправленный HTML
        cdata_section = ET.Element("![CDATA[")
        cdata_section.text = unescape_special_characters(repaired_html)
        cdata_section.tail = "]]>"
        element.text = None
        element.append(cdata_section)

# Сохранение обновленного XML-файла
updated_xml_file_path = 'fileU.xml'
tree.write(updated_xml_file_path, encoding="utf-8", xml_declaration=True)

# Удаление экранирования символов < и >
with open(updated_xml_file_path, 'r', encoding='utf-8') as file:
    xml_content = file.read()
    xml_content = html.unescape(xml_content)
    xml_content = xml_content.replace('& nbsp;', ' ')
    xml_content = xml_content.replace('& nbsp;', ' ')
    xml_content = xml_content.replace('nbsp;', ' ')
    xml_content = xml_content.replace('& NBSP;', ' ')
    xml_content = xml_content.replace('NBSP;', ' ')
    xml_content = xml_content.replace('& Nbsp;', ' ')
    xml_content = xml_content.replace('Nbsp;', ' ')
    xml_content = xml_content.replace('NBSP', ' ')
    xml_content = xml_content.replace('Nbsp;', ' ')
    xml_content = xml_content.replace('<html>', '')
    xml_content = xml_content.replace('<body>', '')
    xml_content = xml_content.replace('</html>', '')
    xml_content = xml_content.replace('</body>', '')
    xml_content = xml_content.replace('&nbsp;', ' ')
    xml_content = xml_content.replace('< /Div>', '</div>')
    xml_content = xml_content.replace('</![CDATA[>', '')
    xml_content = xml_content.replace('<![CDATA[ >', '<![CDATA[')
    xml_content = xml_content.replace('<![CDATA[>', '<![CDATA[')  
    xml_content = xml_content.replace('amp ;', 'amp;')
    xml_content = xml_content.replace('Heckler&Koch', 'Heckler &amp; Koch')
    xml_content = xml_content.replace('Smith&Wesson', 'Smith &amp; Wesson')
    # Удаление пустых строки
    xml_content = '\n'.join(line for line in xml_content.splitlines() if line.strip())
    
with open(updated_xml_file_path, 'w', encoding='utf-8') as file:
    file.write(xml_content)

print("Процесс завершен. XML-файл успешно обновлен.")
