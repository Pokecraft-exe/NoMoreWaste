<?php

function xlsx_col_letter($index) {
    $letter = '';
    $index++;
    while ($index > 0) {
        $rem = ($index - 1) % 26;
        $letter = chr(65 + $rem) . $letter;
        $index = intdiv($index - 1, 26);
    }
    return $letter;
}

function xlsx_escape($value) {
    return htmlspecialchars((string) $value, ENT_XML1 | ENT_COMPAT, 'UTF-8');
}

function build_xlsx($headers, $rows) {
    $allRows = array_merge([$headers], $rows);
    $sheetRows = '';
    foreach (array_values($allRows) as $rowIndex => $row) {
        $rNum = $rowIndex + 1;
        $cells = '';
        foreach (array_values($row) as $colIndex => $value) {
            $ref = xlsx_col_letter($colIndex) . $rNum;
            if (is_int($value) || is_float($value)) {
                $cells .= '<c r="' . $ref . '"><v>' . $value . '</v></c>';
            } else {
                $cells .= '<c r="' . $ref . '" t="inlineStr"><is><t xml:space="preserve">' . xlsx_escape($value) . '</t></is></c>';
            }
        }
        $sheetRows .= '<row r="' . $rNum . '">' . $cells . '</row>';
    }

    $contentTypes = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        . '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">'
        . '<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>'
        . '<Default Extension="xml" ContentType="application/xml"/>'
        . '<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>'
        . '<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>'
        . '</Types>';

    $rootRels = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        . '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
        . '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>'
        . '</Relationships>';

    $workbook = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        . '<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">'
        . '<sheets><sheet name="Planning" sheetId="1" r:id="rId1"/></sheets>'
        . '</workbook>';

    $workbookRels = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        . '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
        . '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>'
        . '</Relationships>';

    $sheet = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        . '<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">'
        . '<sheetData>' . $sheetRows . '</sheetData>'
        . '</worksheet>';

    $tmpFile = tempnam(sys_get_temp_dir(), 'nmwxlsx');
    $zip = new ZipArchive();
    $zip->open($tmpFile, ZipArchive::OVERWRITE);
    $zip->addFromString('[Content_Types].xml', $contentTypes);
    $zip->addFromString('_rels/.rels', $rootRels);
    $zip->addFromString('xl/workbook.xml', $workbook);
    $zip->addFromString('xl/_rels/workbook.xml.rels', $workbookRels);
    $zip->addFromString('xl/worksheets/sheet1.xml', $sheet);
    $zip->close();

    $content = file_get_contents($tmpFile);
    unlink($tmpFile);
    return $content;
}

function send_xlsx($filename, $headers, $rows) {
    $content = build_xlsx($headers, $rows);
    header('Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
    header('Content-Disposition: attachment; filename="' . $filename . '"');
    header('Content-Length: ' . strlen($content));
    echo $content;
    exit;
}
