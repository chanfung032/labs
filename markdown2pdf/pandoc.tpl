% xeTex cjk template for pandoc

\documentclass{article}

$if(title)$
\title{$title$}
$endif$
$if(author)$
\author{$for(author)$$author$$sep$\\$endfor$}
$endif$
\date{}

\usepackage{xeCJK}    

% 字体
\usepackage{fontspec,xunicode}
\setCJKmainfont{WenQuanYi Micro Hei} 
%\setCJKmainfont{FZMiaoWuS-GB}

% 1.5倍行距
\linespread{1.5}

% 尺寸，需在hyperref包之前引用，否则可能出现第一页尺寸不对的问题
\usepackage[paperwidth=9cm,paperheight=11.6cm,margin=0.1cm,nohead,nofoot]{geometry}

% pdf书签，如果文档内容用UTF-8编码，把CJKbookmarks改成unicode
\usepackage[dvipdfmx, colorlinks=false, pdfborder={0 0 1}]{hyperref} 

% 添加pdf描述信息
\makeatletter
\hypersetup{
    pdfauthor={\@author},
    pdftitle={\@title}
}
\makeatother

% 段首缩进两格
\usepackage{indentfirst}
\setlength{\parindent}{2em}

% 删除页码显示
\pagestyle{empty}

\setlength{\parskip}{6pt plus 2pt minus 1pt}
\setcounter{secnumdepth}{0}

$for(header-includes)$
$header-includes$
$endfor$

\begin{document}

% 如果文档内容用UTF-8编码，把GBK改成UTF8
% \begin{CJK*}{UTF8}{gbsn}                          

$if(title)$
\maketitle
$endif$
\thispagestyle{empty}
\clearpage

$for(include-before)$
$include-before$
$endfor$

%$if(toc)$
%\tableofcontents
%$endif$

$body$

$for(include-after)$
$include-after$
$endfor$

% See http://lists.ffii.org/pipermail/cjk/2008-April/002218.html
% "CJKutf8 results in error in TOC": This is documented in CJK.txt, section 'Possible errors'
% UTF-8 编码时打开, 否则第二遍 latex 时报告章节标题错误，感谢 snoopyzhao@newsmth 指出
\clearpage
%\end{CJK*}

\end{document}
