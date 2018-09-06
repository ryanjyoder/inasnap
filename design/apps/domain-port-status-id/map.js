function (doc) {
    var status = "";
    if( doc.status){
        status = doc.status;
    }
    emit([doc.domain,doc.port, status], doc._id);
}