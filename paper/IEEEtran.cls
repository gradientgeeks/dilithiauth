\NeedsTeXFormat{LaTeX2e}
\ProvidesClass{IEEEtran}[2015/08/26 v1.8b by Michael Shell]

\DeclareOption{conference}{\PassOptionsToClass{twocolumn,10pt}{article}}
\DeclareOption{journal}{\PassOptionsToClass{twocolumn,10pt}{article}}
\DeclareOption{technote}{\PassOptionsToClass{twocolumn,9pt}{article}}
\DeclareOption{peerreview}{\PassOptionsToClass{onecolumn,11pt}{article}}
\DeclareOption{draft}{\PassOptionsToClass{draft}{article}}
\DeclareOption{draftcls}{\PassOptionsToClass{draft}{article}}
\DeclareOption{draftclsnofoot}{\PassOptionsToClass{draft}{article}}
\DeclareOption{final}{\PassOptionsToClass{final}{article}}
\DeclareOption{letterpaper}{\PassOptionsToClass{letterpaper}{article}}
\DeclareOption{a4paper}{\PassOptionsToClass{a4paper}{article}}
\DeclareOption{oneside}{\PassOptionsToClass{oneside}{article}}
\DeclareOption{twoside}{\PassOptionsToClass{twoside}{article}}
\DeclareOption{onecolumn}{\PassOptionsToClass{onecolumn}{article}}
\DeclareOption{twocolumn}{\PassOptionsToClass{twocolumn}{article}}
\DeclareOption{10pt}{\PassOptionsToClass{10pt}{article}}
\DeclareOption{11pt}{\PassOptionsToClass{11pt}{article}}
\DeclareOption{12pt}{\PassOptionsToClass{12pt}{article}}
\DeclareOption*{\PassOptionsToClass{\CurrentOption}{article}}

\ExecuteOptions{letterpaper,10pt,twocolumn,final}
\ProcessOptions\relax

\LoadClass{article}

\RequirePackage{geometry}
\geometry{letterpaper, top=0.75in, bottom=1.0in, left=0.625in, right=0.625in}

\newcommand{\IEEEauthorblockN}[1]{{\large #1}\\[0.3em]}
\newcommand{\IEEEauthorblockA}[1]{\parbox[t]{0.8\textwidth}{\centering\small #1}}

\newenvironment{IEEEkeywords}{%
  \vspace{0.5em}\noindent\textbf{Index Terms---}\ignorespaces
}{%
  \par\vspace{1em}
}

\renewcommand{\abstractname}{\textbf{Abstract}}
