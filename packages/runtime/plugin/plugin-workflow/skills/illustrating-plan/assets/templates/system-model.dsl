/*
 Fill in: copy to plans/<change>/diagrams/<topic>-model.dsl, replace every <placeholder> and delete the comment blocks.
 Use it only when several views share the elements. Render command, from the diagrams directory:
   docker run --rm -v "$PWD:/usr/local/structurizr" structurizr/structurizr export -workspace /usr/local/structurizr/<topic>-model.dsl -format plantuml -output /usr/local/structurizr
 Then render each exported <topic>-view.puml to SVG, or to PNG only when the user asked for PNG.
*/
workspace "<Change name>" "<One sentence stating what the model shows>" {

    model {
        user = person "<Role of the user>" "<What the user does>"

        /* Fill in: tag every element and relationship with its state: Settled, Open, Future or Current. */
        system = softwareSystem "<System that changes>" "<What it does>" "Settled" {
            api = container "<Component A>" "<What it does>" "<Technology>" "Settled"
            sender = container "<Component B>" "<What it does>" "<Technology>" "Open"
        }
        legacy = softwareSystem "<System being retired>" "<What it does today>" "Current"

        user -> api "<Settled call>" "<Protocol>" "Settled"
        api -> sender "<Call that depends on an open decision>" "<Protocol>" "Open"
        api -> legacy "<Call that goes away>" "<Protocol>" "Current"
    }

    views {
        /* Fill in: one view for each question, keyed <topic>-view. The settled view excludes everything open. */
        container system "<topic>-view" "<The question this view answers>" {
            include *
            exclude "element.tag==Open"
            autoLayout tb
        }

        styles {
            element "Settled" {
                background #C8E6C9
            }
            element "Open" {
                background #FFE0B2
            }
            element "Future" {
                background #BBDEFB
            }
            element "Current" {
                background #E0E0E0
            }
        }
    }
}
