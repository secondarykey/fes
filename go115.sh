NOW=`pwd`
DIR="app/handler"

INTERNAL_DIR=${NOW}/${DIR}/internal

if [ -e ${INTERNAL_DIR}/static.go115 ]; then

  echo "switch template"
  mv ${INTERNAL_DIR}/template.go ${INTERNAL_DIR}/template.go116
  mv ${INTERNAL_DIR}/template.go115 ${INTERNAL_DIR}/template.go

  echo "switch static"
  mv ${INTERNAL_DIR}/static.go ${INTERNAL_DIR}/static.go116
  mv ${INTERNAL_DIR}/static.go115 ${INTERNAL_DIR}/static.go

  echo "switch archive"
  mv ${INTERNAL_DIR}/archive.go ${INTERNAL_DIR}/archive.go116
  mv ${INTERNAL_DIR}/archive.go115 ${INTERNAL_DIR}/archive.go

  echo "generate statik"
  cd ${INTERNAL_DIR}
  statik -src _assets

  cd ${NOW}

  echo "Success"
else
  echo "Now Version 1.15???"
fi

